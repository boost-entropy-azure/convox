package api

import "github.com/convox/stdapi"

func (s *Server) setupRoutes(r stdapi.Router) {
	r.Use(s.Authorize)

	r.Route("POST", "/apps/{name}/cancel", s.AppCancel)
	r.Route("POST", "/apps", s.AppCreate)
	r.Route("DELETE", "/apps/{name}", s.AppDelete)
	r.Route("GET", "/apps/{name}", s.AppGet)
	r.Route("GET", "/apps", s.AppList)
	r.Route("SOCKET", "/apps/{name}/logs", s.AppLogs)
	r.Route("GET", "/apps/{name}/metrics", s.AppMetrics)
	r.Route("PUT", "/apps/{name}", s.AppUpdate)
	r.Route("GET", "/apps/{app}/balancers", s.BalancerList)
	r.Route("POST", "/apps/{app}/builds", s.BuildCreate)
	r.Route("GET", "/apps/{app}/builds/{id}.tgz", s.BuildExport)
	r.Route("GET", "/apps/{app}/builds/{id}", s.BuildGet)
	r.Route("POST", "/apps/{app}/builds/import", s.BuildImport)
	r.Route("GET", "/apps/{app}/builds", s.BuildList)
	r.Route("SOCKET", "/apps/{app}/builds/{id}/logs", s.BuildLogs)
	r.Route("PUT", "/apps/{app}/builds/{id}", s.BuildUpdate)
	r.Route("GET", "/apps/{app}/configs", s.AppConfigList)
	r.Route("GET", "/apps/{app}/configs/{name}", s.AppConfigGet)
	r.Route("PUT", "/apps/{app}/configs/{name}", s.AppConfigSet)
	r.Route("GET", "/system/capacity", s.CapacityGet)
	r.Route("PUT", "/apps/{app}/ssl/{service}/{port}", s.CertificateApply)
	r.Route("POST", "/certificates", s.CertificateCreate)
	r.Route("DELETE", "/certificates/{id}", s.CertificateDelete)
	r.Route("POST", "/certificates/{id}/renew", s.CertificateRenew)
	r.Route("POST", "/certificates/generate", s.CertificateGenerate)
	r.Route("GET", "/certificates", s.CertificateList)
	r.Route("GET", "/letsencrypt/config", s.LetsEncryptConfigGet)
	r.Route("PUT", "/letsencrypt/config", s.LetsEncryptConfigApply)
	r.Route("POST", "/events", s.EventSend)
	r.Route("DELETE", "/apps/{app}/processes/{pid}/files", s.FilesDelete)
	r.Route("GET", "/apps/{app}/processes/{pid}/files", s.FilesDownload)
	r.Route("POST", "/apps/{app}/processes/{pid}/files", s.FilesUpload)
	r.Route("", "", s.Initialize)
	r.Route("POST", "/instances/keyroll", s.InstanceKeyroll)
	r.Route("GET", "/instances", s.InstanceList)
	r.Route("SOCKET", "/instances/{id}/shell", s.InstanceShell)
	r.Route("DELETE", "/instances/{id}", s.InstanceTerminate)
	r.Route("DELETE", "/apps/{app}/objects/{key:.*}", s.ObjectDelete)
	r.Route("HEAD", "/apps/{app}/objects/{key:.*}", s.ObjectExists)
	r.Route("GET", "/apps/{app}/objects/{key:.*}", s.ObjectFetch)
	r.Route("GET", "/apps/{app}/objects", s.ObjectList)
	r.Route("POST", "/apps/{app}/objects/{key:.*}", s.ObjectStore)
	r.Route("SOCKET", "/apps/{app}/processes/{pid}/exec", s.ProcessExec)
	r.Route("GET", "/apps/{app}/processes/{pid}", s.ProcessGet)
	r.Route("GET", "/apps/{app}/processes", s.ProcessList)
	r.Route("SOCKET", "/apps/{app}/processes/{pid}/logs", s.ProcessLogs)
	r.Route("POST", "/apps/{app}/services/{service}/processes", s.ProcessRun)
	r.Route("SOCKET", "/apps/{app}/services/{service}/logs", s.ServiceLogs)
	r.Route("DELETE", "/apps/{app}/processes/{pid}", s.ProcessStop)
	r.Route("SOCKET", "/proxy/{host}/{port}", s.Proxy)
	r.Route("POST", "/registries", s.RegistryAdd)
	r.Route("GET", "/registries", s.RegistryList)
	r.Route("ANY", "/v2/{path:.*}", s.RegistryProxy)
	r.Route("DELETE", "/registries/{server:.*}", s.RegistryRemove)
	r.Route("POST", "/apps/{app}/releases", s.ReleaseCreate)
	r.Route("GET", "/apps/{app}/releases/{id}", s.ReleaseGet)
	r.Route("GET", "/apps/{app}/releases", s.ReleaseList)
	r.Route("POST", "/apps/{app}/releases/{id}/promote", s.ReleasePromote)
	r.Route("SOCKET", "/apps/{app}/resources/{name}/console", s.ResourceConsole)
	r.Route("GET", "/apps/{app}/resources/{name}/data", s.ResourceExport)
	r.Route("GET", "/apps/{app}/resources/{name}", s.ResourceGet)
	r.Route("PUT", "/apps/{app}/resources/{name}/data", s.ResourceImport)
	r.Route("GET", "/apps/{app}/resources", s.ResourceList)
	r.Route("GET", "/apps/{app}/services", s.ServiceList)
	r.Route("POST", "/apps/{app}/services/{name}/restart", s.ServiceRestart)
	r.Route("PUT", "/apps/{app}/services/{name}", s.ServiceUpdate)
	r.Route("", "", s.Start)
	r.Route("GET", "/system", s.SystemGet)
	r.Route("", "", s.SystemInstall)
	r.Route("SOCKET", "/system/logs", s.SystemLogs)
	r.Route("GET", "/system/metrics", s.SystemMetrics)
	r.Route("PUT", "/system/jwt/rotate", s.SystemJwtSignKeyRotate)
	r.Route("POST", "/system/jwt/token", s.SystemJwtToken)
	r.Route("GET", "/system/processes", s.SystemProcesses)
	r.Route("GET", "/system/releases", s.SystemReleases)
	r.Route("POST", "/resources", s.SystemResourceCreate)
	r.Route("DELETE", "/resources/{name}", s.SystemResourceDelete)
	r.Route("GET", "/resources/{name}", s.SystemResourceGet)
	r.Route("POST", "/resources/{name}/links", s.SystemResourceLink)
	r.Route("GET", "/resources", s.SystemResourceList)
	r.Route("OPTIONS", "/resources", s.SystemResourceTypes)
	r.Route("DELETE", "/resources/{name}/links/{app}", s.SystemResourceUnlink)
	r.Route("PUT", "/resources/{name}", s.SystemResourceUpdate)
	r.Route("", "", s.SystemUninstall)
	r.Route("PUT", "/system", s.SystemUpdate)
	r.Route("", "", s.Workers)

	r.Route("ANY", "/custom/http/proxy/{path:.*}", s.ProxyHttpService)
}
