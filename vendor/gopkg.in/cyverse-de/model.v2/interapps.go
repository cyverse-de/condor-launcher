package model

// InteractiveApps contains the settings needed for interactive apps across all
// steps in a Job.
type InteractiveApps struct {
	// The docker image for the reverse proxy that runs on the cluster with the
	// job steps.
	ProxyImage string `json:"proxy_image"`

	// The name of the container for the reverse proxy.
	ProxyName string `json:"proxy_name"`

	// The URL for the frontend of the application. Will get prefixed with the job
	// id.
	FrontendURL string `json:"frontend_url"`

	// The base URL for the CAS server.
	CASURL string `json:"cas_url"`

	// The path to the validate endpoint on the CAS server.
	CASValidate string `json:"cas_validate"`

	// The path to the SSL cert file on the Condor nodes.
	SSLCertPath string `json:"ssl_cert_path"`

	// The path to the SSL key file on the Condor nodes.
	SSLKeyPath string `json:"ssl_key_path"`

	// If websocket handling requires a special path in the app. The default is to
	// have this be empty.
	WebsocketPath string `json:"websocket_path"`

	// If websocket handling requires a special port in the app. The default is to
	// use the same port as the backend URL.
	WebsocketPort string `json:"websocket_port"`

	// If websocket handling requires a protocol other than ws://.
	WebsocketProto string `json:"websocket_proto"`

	// Only used if you need to override the default backendURL, which should be
	// http://<container_name>.
	BackendURL string `json:"backend_url"`
}
