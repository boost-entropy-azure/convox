package cli

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/convox/convox/pkg/common"
	"github.com/convox/convox/pkg/manifest"
	"github.com/convox/convox/pkg/options"
	"github.com/convox/convox/pkg/rack"
	"github.com/convox/convox/pkg/structs"
	"github.com/convox/convox/provider"
	"github.com/convox/convox/sdk"
	"github.com/convox/stdcli"
)

func init() {
	register("rack", "get information about the rack", watch(Rack), stdcli.CommandOptions{
		Flags:    []stdcli.Flag{flagRack, flagWatchInterval},
		Validate: stdcli.Args(0),
	})

	register("rack access", "get rack access credential", RackAccess, stdcli.CommandOptions{
		Flags: []stdcli.Flag{
			flagRack,
			stdcli.StringFlag("role", "", "access role: read or write"),
			stdcli.IntFlag("duration-in-hours", "", "duration in hours"),
		},
		Validate: stdcli.Args(0),
	})

	register("rack access key rotate", "rotate access key", RackAccessKeyRotate, stdcli.CommandOptions{
		Flags:    []stdcli.Flag{flagRack},
		Validate: stdcli.Args(0),
	})

	registerWithoutProvider("rack install", "install a new rack", RackInstall, stdcli.CommandOptions{
		Flags: []stdcli.Flag{
			stdcli.BoolFlag("prepare", "", "prepare the install but don't run it"),
			stdcli.StringFlag("version", "v", "rack version"),
			stdcli.StringFlag("runtime", "r", "runtime id"),
		},
		Usage:    "<provider> <name> [option=value]...",
		Validate: stdcli.ArgsMin(2),
	})

	registerWithoutProvider("rack kubeconfig", "generate kubeconfig for rack", RackKubeconfig, stdcli.CommandOptions{
		Flags:    []stdcli.Flag{flagRack},
		Validate: stdcli.Args(0),
	})

	register("rack logs", "get logs for the rack", RackLogs, stdcli.CommandOptions{
		Flags:    append(stdcli.OptionFlags(structs.LogsOptions{}), flagNoFollow, flagRack),
		Validate: stdcli.Args(0),
	})

	registerWithoutProvider("rack mv", "move a rack to or from console", RackMv, stdcli.CommandOptions{
		Usage:    "<from> <to>",
		Validate: stdcli.Args(2),
	})

	registerWithoutProvider("rack params", "display rack parameters", RackParams, stdcli.CommandOptions{
		Flags:    []stdcli.Flag{flagRack},
		Validate: stdcli.Args(0),
	})

	registerWithoutProvider("rack params set", "set rack parameters", RackParamsSet, stdcli.CommandOptions{
		Flags:    []stdcli.Flag{flagRack},
		Usage:    "<Key=Value> [Key=Value]...",
		Validate: stdcli.ArgsMin(1),
	})

	register("rack ps", "list rack processes", RackPs, stdcli.CommandOptions{
		Flags:    append(stdcli.OptionFlags(structs.SystemProcessesOptions{}), flagRack),
		Validate: stdcli.Args(0),
	})

	register("rack releases", "list rack version history", RackReleases, stdcli.CommandOptions{
		Flags:    []stdcli.Flag{flagRack},
		Validate: stdcli.Args(0),
	})

	register("rack runtimes", "list attachable runtime integrations", RackRuntimes, stdcli.CommandOptions{
		Flags:    []stdcli.Flag{flagRack},
		Validate: stdcli.Args(0),
	})

	register("rack runtime attach", "attach runtime integration", RackRuntimeAttach, stdcli.CommandOptions{
		Flags:    []stdcli.Flag{flagRack},
		Validate: stdcli.Args(1),
	})

	register("rack scale", "scale the rack", RackScale, stdcli.CommandOptions{
		Flags: []stdcli.Flag{
			flagRack,
			stdcli.IntFlag("count", "c", "instance count"),
			stdcli.StringFlag("type", "t", "instance type"),
		},
		Validate: stdcli.Args(0),
	})

	register("rack sync", "sync v2 rack API url", RackSync, stdcli.CommandOptions{
		Flags:    []stdcli.Flag{flagRack},
		Validate: stdcli.Args(0),
	})

	registerWithoutProvider("rack uninstall", "uninstall a rack", RackUninstall, stdcli.CommandOptions{
		Usage:    "<name>",
		Validate: stdcli.Args(1),
	})

	registerWithoutProvider("rack update", "update a rack", RackUpdate, stdcli.CommandOptions{
		Flags:    []stdcli.Flag{flagRack, flagForce},
		Usage:    "[version]",
		Validate: stdcli.ArgsMax(1),
	})
}

type NodeGroupConfigParam struct {
	Id           *int    `json:"id"`
	Type         string  `json:"type"`
	Disk         *int    `json:"disk,omitempty"`
	CapacityType *string `json:"capacity_type,omitempty"`
	MinSize      *int    `json:"min_size,omitempty"`
	MaxSize      *int    `json:"max_size,omitempty"`
	Label        *string `json:"label,omitempty"`
	AmiID        *string `json:"ami_id,omitempty"`
	Dedicated    *bool   `json:"dedicated,omitempty"`
	Tags         *string `json:"tags,omitempty"`
}

func (n *NodeGroupConfigParam) Validate() error {
	if n.Type == "" {
		return fmt.Errorf("node type is required: '%s'", n.Type)
	}
	if n.Disk != nil && *n.Disk < 20 {
		return fmt.Errorf("node disk is less than 20: '%d'", *n.Disk)
	}
	if n.MinSize != nil && *n.MinSize < 0 {
		return fmt.Errorf("invalid min size: '%d'", *n.MinSize)
	}
	if n.MaxSize != nil && *n.MaxSize < 0 {
		return fmt.Errorf("invalid max size: '%d'", *n.MaxSize)
	}
	if n.MinSize != nil && n.MaxSize != nil && *n.MinSize > *n.MaxSize {
		return fmt.Errorf("invalid min size: '%d' must be less or equal to max size", *n.MinSize)
	}
	if n.CapacityType != nil && (*n.CapacityType != "ON_DEMAND" && *n.CapacityType != "SPOT") {
		return fmt.Errorf("allowed capasity type: ON_DEMAND or SPOT, found : '%s'", *n.CapacityType)
	}
	if n.Label != nil && !manifest.NameValidator.MatchString(*n.Label) {
		return fmt.Errorf("label value '%s' invalid, %s", *n.Label, manifest.ValidNameDescription)
	}

	if n.Dedicated != nil && *n.Dedicated && n.Label == nil {
		return fmt.Errorf("label is required when dedicated option is set")
	}

	if n.Tags != nil {
		reserved := []string{"name", "rack"}
		for _, part := range strings.Split(*n.Tags, ",") {
			if len(strings.SplitN(part, "=", 2)) != 2 {
				return fmt.Errorf("invalid 'tags', use format: k1=v1,k2=v2")
			}

			k := strings.SplitN(part, "=", 2)[0]
			if common.ContainsInStringSlice(reserved, strings.ToLower(k)) {
				return fmt.Errorf("reserved tag key '%s' is not allowed", k)
			}
		}
	}

	return nil
}

type AdditionalNodeGroups []NodeGroupConfigParam

func (an AdditionalNodeGroups) Validate() error {
	idCnt := 0
	idMap := map[int]bool{}
	for i := range an {
		if err := an[i].Validate(); err != nil {
			return err
		}
		if an[i].Id != nil {
			idCnt++
			if idMap[*an[i].Id] {
				return fmt.Errorf("duplicate node group id is found: %d", *an[i].Id)
			}
		}
	}

	if idCnt > 0 && idCnt != len(an) {
		return fmt.Errorf("some node groups missing id property")
	}

	if idCnt == 0 {
		for i := range an {
			an[i].Id = options.Int(i + 1)
		}
	}
	return nil
}

func validateAndMutateParams(params map[string]string) error {
	if params["high_availability"] != "" {
		return errors.New("the high_availability parameter is only supported during rack installation")
	}

	srdown, srup := params["ScheduleRackScaleDown"], params["ScheduleRackScaleUp"]
	if (srdown == "" || srup == "") && (srdown != "" || srup != "") {
		return errors.New("to schedule your rack to turn on/off you need both ScheduleRackScaleDown and ScheduleRackScaleUp parameters")
	}

	// format: "key1=val1,key2=val2"
	if tags, has := params["tags"]; has {
		tList := strings.Split(tags, ",")
		for _, p := range tList {
			if len(strings.Split(p, "=")) != 2 {
				return errors.New("invalid value for tags param")
			}
		}
	}

	ngKeys := []string{"additional_node_groups_config", "additional_build_groups_config"}
	for _, k := range ngKeys {
		if params[k] != "" {
			var err error
			cfgData := []byte(params[k])
			if strings.HasSuffix(params[k], ".json") {
				cfgData, err = os.ReadFile(params[k])
				if err != nil {
					return fmt.Errorf("invalid param '%s' value, failed to read the file: %s", k, err)
				}
			} else if !strings.HasPrefix(params[k], "[") {
				data, err := base64.StdEncoding.DecodeString(params[k])
				if err != nil {
					return fmt.Errorf("invalid param '%s' value: %s", k, err)
				}

				cfgData = data
			}

			nCfgs := AdditionalNodeGroups{}
			if err := json.Unmarshal(cfgData, &nCfgs); err != nil {
				return err
			}

			if err := nCfgs.Validate(); err != nil {
				return err
			}

			sort.Slice(nCfgs, func(i, j int) bool {
				if nCfgs[i].Id == nil || nCfgs[j].Id == nil {
					return true
				}
				return *nCfgs[i].Id < *nCfgs[j].Id
			})

			data, err := json.Marshal(nCfgs)
			if err != nil {
				return fmt.Errorf("failed to process params '%s': %s", k, err)
			}
			params[k] = base64.StdEncoding.EncodeToString(data)
		}
	}

	return nil
}

func Rack(rack sdk.Interface, c *stdcli.Context) error {
	s, err := rack.SystemGet()
	if err != nil {
		return err
	}

	i := c.Info()

	i.Add("Name", s.Name)
	i.Add("Provider", s.Provider)

	if s.Region != "" {
		i.Add("Region", s.Region)
	}

	if s.Domain != "" {
		if ri := s.Outputs["DomainInternal"]; ri != "" {
			i.Add("Router", fmt.Sprintf("%s (external)\n%s (internal)", s.Domain, ri))
		} else {
			i.Add("Router", s.Domain)
		}
	}

	if s.RouterInternal != "" {
		i.Add("RouterInternal", s.RouterInternal)
	}

	i.Add("Status", s.Status)
	i.Add("Version", s.Version)

	return i.Print()
}

func RackAccess(rack sdk.Interface, c *stdcli.Context) error {
	data, err := c.SettingRead("current")
	if err != nil {
		return err
	}
	var attrs map[string]string
	if err := json.Unmarshal([]byte(data), &attrs); err != nil {
		return err
	}

	rData, err := rack.SystemGet()
	if err != nil {
		return err
	}

	role, ok := c.Value("role").(string)
	if !ok {
		return fmt.Errorf("role is required")
	}

	duration, ok := c.Value("duration-in-hours").(int)
	if !ok {
		return fmt.Errorf("duration is required")
	}

	jwtTk, err := rack.SystemJwtToken(structs.SystemJwtOptions{
		Role:           options.String(role),
		DurationInHour: options.String(strconv.Itoa(duration)),
	})
	if err != nil {
		return err
	}

	return c.Writef("RACK_URL=https://jwt:%s@%s\n", jwtTk.Token, rData.RackDomain)
}

func RackAccessKeyRotate(rack sdk.Interface, c *stdcli.Context) error {
	_, err := rack.SystemJwtSignKeyRotate()
	if err != nil {
		return err
	}

	return c.OK()
}

func RackInstall(_ sdk.Interface, c *stdcli.Context) error {
	slug := c.Arg(0)
	name := c.Arg(1)
	args := c.Args[2:]
	version := c.String("version")
	runtime := c.String("runtime")

	if !provider.Valid(slug) {
		return fmt.Errorf("unknown provider: %s", slug)
	}
	var parts []string
	if runtime != "" {
		parts = strings.Split(name, "/")
		name = parts[1]
	}

	if err := checkRackNameRegex(name); err != nil {
		return err
	}

	opts := argsToOptions(args)

	if c.Bool("prepare") {
		opts["release"] = version

		md := &rack.Metadata{
			Provider: slug,
			Vars:     opts,
		}

		if _, err := rack.Create(c, name, md); err != nil {
			return err
		}

		return nil
	}

	if runtime != "" {
		name = parts[0] + "/" + parts[1]
	}

	if err := rack.Install(c, slug, name, version, runtime, opts); err != nil {
		return err
	}

	if runtime != "" {
		c.Writef("Convox Rack installation initiated. Check the progress on the Console Racks page if desired. \n")
	}

	if _, err := rack.Current(c); err != nil {
		if _, err := rack.Switch(c, name); err != nil {
			return err
		}
	}

	return nil
}

func RackKubeconfig(_ sdk.Interface, c *stdcli.Context) error {
	r, err := rack.Current(c)
	if err != nil {
		return err
	}

	ep, err := r.Endpoint()
	if err != nil {
		return err
	}

	pw, _ := ep.User.Password()

	data := strings.TrimSpace(fmt.Sprintf(`
apiVersion: v1
clusters:
- cluster:
    server: %s://%s/kubernetes/
  name: rack
contexts:
- context:
    cluster: rack
    user: convox
  name: convox@rack
current-context: convox@rack
kind: Config
users:
- name: convox
  user:
    username: "%s"
    password: "%s"
	`, ep.Scheme, ep.Host, ep.User.Username(), pw))

	fmt.Println(data)

	return nil
}

func RackLogs(rack sdk.Interface, c *stdcli.Context) error {
	var opts structs.LogsOptions

	if err := c.Options(&opts); err != nil {
		return err
	}

	if c.Bool("no-follow") {
		opts.Follow = options.Bool(false)
	}

	opts.Prefix = options.Bool(true)

	r, err := rack.SystemLogs(opts)
	if err != nil {
		return err
	}

	io.Copy(c, r)

	return nil
}

func RackMv(_ sdk.Interface, c *stdcli.Context) error {
	from := c.Arg(0)
	to := c.Arg(1)

	movedToConsole, toRackName := false, to
	parts := strings.SplitN(to, "/", 2)
	if len(parts) == 2 {
		movedToConsole = true
		toRackName = parts[1]
	}

	fromRackName := from
	fparts := strings.SplitN(from, "/", 2)
	if len(fparts) == 2 {
		fromRackName = fparts[1]
	}

	if fromRackName != toRackName {
		return fmt.Errorf("rack name must remain same")
	}

	c.Startf("moving rack <rack>%s</rack> to <rack>%s</rack>", from, to)

	fr, err := rack.Load(c, from)
	if err != nil {
		return err
	}

	md, err := fr.Metadata()
	if err != nil {
		return err
	}

	if !md.Deletable {
		return fmt.Errorf("rack %s has dependencies and can not be moved", from)
	}

	md, err = fr.Metadata()
	if err != nil {
		return err
	}

	if _, err := rack.Create(c, to, md); err != nil {
		return err
	}

	if err := fr.Delete(); err != nil {
		return err
	}

	if movedToConsole {
		ci := c.Info()
		ci.Add("Attention!", "Login in the console and attach a runtime integration to the rack")
	}

	return c.OK()
}

func RackParams(_ sdk.Interface, c *stdcli.Context) error {
	r, err := rack.Current(c)
	if err != nil {
		return err
	}

	params, err := r.Parameters()
	if err != nil {
		return err
	}

	keys := []string{}

	for k := range params {
		keys = append(keys, k)
	}

	ngKeys := []string{"additional_node_groups_config", "additional_build_groups_config"}
	for _, k := range ngKeys {
		if params[k] != "" {
			v, err := base64.StdEncoding.DecodeString(params[k])
			if err == nil {
				params[k] = string(v)
			}
		}
	}

	sort.Strings(keys)

	i := c.Info()

	for _, k := range keys {
		i.Add(k, params[k])
	}

	return i.Print()
}

func RackParamsSet(_ sdk.Interface, c *stdcli.Context) error {
	r, err := rack.Current(c)
	if err != nil {
		return err
	}

	c.Startf("Updating parameters")

	params := argsToOptions(c.Args)
	if err := validateAndMutateParams(params); err != nil {
		return err
	}

	if err := r.UpdateParams(params); err != nil {
		return err
	}

	return c.OK()
}

func RackPs(rack sdk.Interface, c *stdcli.Context) error {
	var opts structs.SystemProcessesOptions

	if err := c.Options(&opts); err != nil {
		return err
	}

	ps, err := rack.SystemProcesses(opts)
	if err != nil {
		return err
	}

	t := c.Table("ID", "APP", "SERVICE", "STATUS", "RELEASE", "STARTED", "COMMAND")

	for _, p := range ps {
		t.AddRow(p.Id, p.App, p.Name, p.Status, p.Release, common.Ago(p.Started), p.Command)
	}

	return t.Print()
}

func RackReleases(rack sdk.Interface, c *stdcli.Context) error {
	rs, err := rack.SystemReleases()
	if err != nil {
		return err
	}

	t := c.Table("VERSION", "UPDATED")

	for _, r := range rs {
		t.AddRow(r.Id, common.Ago(r.Created))
	}

	return t.Print()
}

func RackRuntimes(rack sdk.Interface, c *stdcli.Context) error {
	data, err := c.SettingRead("current")
	if err != nil {
		return err
	}
	var attrs map[string]string
	if err := json.Unmarshal([]byte(data), &attrs); err != nil {
		return err
	}

	rs, err := rack.Runtimes(attrs["name"])
	if err != nil {
		return err
	}

	t := c.Table("ID", "TITLE")
	for _, r := range rs {
		t.AddRow(r.Id, r.Title)
	}

	return t.Print()
}

func RackRuntimeAttach(rack sdk.Interface, c *stdcli.Context) error {
	data, err := c.SettingRead("current")
	if err != nil {
		return err
	}
	var attrs map[string]string
	if err := json.Unmarshal([]byte(data), &attrs); err != nil {
		return err
	}

	if err := rack.RuntimeAttach(attrs["name"], structs.RuntimeAttachOptions{
		Runtime: aws.String(c.Arg(0)),
	}); err != nil {
		return err
	}

	return c.OK()
}

func RackScale(rack sdk.Interface, c *stdcli.Context) error {
	s, err := rack.SystemGet()
	if err != nil {
		return err
	}

	var opts structs.SystemUpdateOptions
	update := false

	if v, ok := c.Value("count").(int); ok {
		opts.Count = options.Int(v)
		update = true
	}

	if v, ok := c.Value("type").(string); ok {
		opts.Type = options.String(v)
		update = true
	}

	if update {
		c.Startf("Scaling rack")

		if err := rack.SystemUpdate(opts); err != nil {
			return err
		}

		return c.OK()
	}

	i := c.Info()

	i.Add("Autoscale", s.Parameters["Autoscale"])
	i.Add("Count", fmt.Sprintf("%d", s.Count))
	i.Add("Status", s.Status)
	i.Add("Type", s.Type)

	return i.Print()
}

func RackSync(_ sdk.Interface, c *stdcli.Context) error {
	r, err := rack.Current(c)
	if err != nil {
		return err
	}

	data, err := c.SettingRead("current")
	if err != nil {
		return err
	}
	var attrs map[string]string
	if err := json.Unmarshal([]byte(data), &attrs); err != nil {
		return err
	}

	if attrs["type"] == "console" {
		m, err := r.Metadata()
		if err != nil {
			return err
		}

		if m.State == nil { // v2 racks don't have a state file
			err := r.Sync()
			if err != nil {
				return err
			}

			return c.OK()
		}
	}

	return fmt.Errorf("sync is only supported for console managed v2 racks")
}

func RackUninstall(_ sdk.Interface, c *stdcli.Context) error {
	name := c.Arg(0)

	r, err := rack.Match(c, name)
	if err != nil {
		return err
	}

	if err := r.Uninstall(); err != nil {
		return err
	}

	return nil
}

func RackUpdate(_ sdk.Interface, c *stdcli.Context) error {
	r, err := rack.Current(c)
	if err != nil {
		return err
	}

	cl, err := r.Client()
	if err != nil {
		return err
	}

	s, err := cl.SystemGet()
	if err != nil {
		return err
	}

	currentVersion := s.Version
	newVersion := c.Arg(0)

	// disable downgrabe from minor version for v3 rack
	if strings.HasPrefix(currentVersion, "3.") && strings.HasPrefix(newVersion, "3.") &&
		!strings.Contains(currentVersion, "rc") && !strings.Contains(newVersion, "rc") {
		curv, err := strconv.Atoi(strings.Split(currentVersion, ".")[1])
		if err != nil {
			return err
		}

		newv, err := strconv.Atoi(strings.Split(newVersion, ".")[1])
		if err != nil {
			return err
		}
		if newv < curv {
			return fmt.Errorf("Downgrade from minor version is not supported for v3 rack. Contact the support.")
		}
	}

	force, _ := c.Value("force").(bool)
	if err := r.UpdateVersion(newVersion, force); err != nil {
		return err
	}

	return nil
}
