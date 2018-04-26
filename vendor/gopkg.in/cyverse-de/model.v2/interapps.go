package model

// InteractiveApps contains the settings needed for interactive apps across all
// steps in a Job.
type InteractiveApps struct {
	ProxyImage  string `json:"proxy_image"`   //The docker image for the reverse proxy that runs on the cluster with the job steps.
	ProxyName   string `json:"proxy_name"`    //The name of the container for the reverse proxy.
	FrontendURL string `json:"frontend_url"`  //The URL for the frontend of the application. Will get prefixed with the job id.
	CASURL      string `json:"cas_url"`       //The base URL for the CAS server.
	CASValidate string `json:"cas_validate"`  //The path to the validate endpoint on the CAS server.
	SSLCertPath string `json:"ssl_cert_path"` //The path to the SSL cert file on the Condor nodes.
	SSLKeyPath  string `json:"ssl_key_path"`  //The path to the SSL key file on the Condor nodes.
}
