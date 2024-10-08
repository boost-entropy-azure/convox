package structs

import (
	"strings"
	"time"

	"github.com/gobwas/glob"
)

type Certificate struct {
	Arn        string    `json:"-"`
	Id         string    `json:"id"`
	Domain     string    `json:"domain"`
	Domains    []string  `json:"domains"`
	Expiration time.Time `json:"expiration"`
	Status     string    `json:"status"`
}

type Certificates []Certificate

type CertificateCreateOptions struct {
	Chain *string `param:"chain"`
}

type CertificateGenerateOptions struct {
	Issuer   *string `flag:"issuer" param:"issuer"`
	Duration *string `flag:"duration" param:"duration"`
}

type CertificateListOptions struct {
	Generated *bool `flag:"generated" param:"generated"`
}

func (c Certificates) Less(i, j int) bool { return strings.ToUpper(c[i].Id) < strings.ToUpper(c[j].Id) }

func (c *Certificate) Match(domain string) (bool, error) {
	for _, d := range c.Domains {
		g, err := glob.Compile(d, '.')
		if err != nil {
			return false, err
		}

		if g.Match(domain) {
			return true, nil
		}
	}

	return false, nil
}
