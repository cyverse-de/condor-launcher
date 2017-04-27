package vaulter

import (
	vault "github.com/hashicorp/vault/api"
)

// Vaulter defines the lower-level interactions with vault so that they can be
// stubbed out in unit tests.
type Vaulter interface {
	Mounter
	MountConfigGetter
	MountLister
	MountWriter
	MountReader
	PathDeleter
	Revoker
}

// VaultAPI provides an implementation of the Vaulter interface that can
// actually hit the Vault API.
type VaultAPI struct {
	client *vault.Client
	cfg    *vault.Config
}

// Token returns a new Vault token.
func (v *VaultAPI) Token() *vault.TokenAuth {
	return v.client.Auth().Token()
}

// CreateToken returns a new child or orphan token.
func (v *VaultAPI) CreateToken(ta *vault.TokenAuth, opts *vault.TokenCreateRequest) (*vault.Secret, error) {
	return ta.Create(opts)
}

// Mount uses the Vault API to mount a backend at a path.
func (v *VaultAPI) Mount(path string, mi *vault.MountInput) error {
	sys := v.client.Sys()
	return sys.Mount(path, mi)
}

// Unmount uses the Vault API to unmount a backend at the provided path.
func (v *VaultAPI) Unmount(path string) error {
	return v.client.Sys().Unmount(path)
}

// MountConfig uses the VaultAPI to get the config for the passed in mount
// point.
func (v *VaultAPI) MountConfig(path string) (*vault.MountConfigOutput, error) {
	sys := v.client.Sys()
	return sys.MountConfig(path)
}

// TuneMount uses the VaultAPI to set the config for the passed in mount
// point.
func (v *VaultAPI) TuneMount(path string, in vault.MountConfigInput) error {
	sys := v.client.Sys()
	return sys.TuneMount(path, in)
}

// ListMounts lists the mounted Vault backends.
func (v *VaultAPI) ListMounts() (map[string]*vault.MountOutput, error) {
	sys := v.client.Sys()
	return sys.ListMounts()
}

// DefaultConfig returns a *vault.Config filled out with the default values.
// They're not just the Go zero values for data types.
func (v *VaultAPI) DefaultConfig() *vault.Config {
	return vault.DefaultConfig()
}

// ConfigureTLS sets up the passed in Vault config for TLS protected
// communication with Vault.
func (v *VaultAPI) ConfigureTLS(cfg *vault.Config, t *vault.TLSConfig) error {
	return cfg.ConfigureTLS(t)
}

// NewClient creates a new Vault client.
func (v *VaultAPI) NewClient(cfg *vault.Config) (*vault.Client, error) {
	return vault.NewClient(cfg)
}

// SetClient sets the value of the internal *vault.Client field.
func (v *VaultAPI) SetClient(c *vault.Client) {
	v.client = c
}

// Client gets the currently configured Vault client.
func (v *VaultAPI) Client() *vault.Client {
	return v.client
}

// GetConfig returns the *vault.Config instance used with the underlying client.
func (v *VaultAPI) GetConfig() *vault.Config {
	return v.cfg
}

// SetConfig sets the vault config that should be used with the underlying
// client. Is NOT called by NewClient().
func (v *VaultAPI) SetConfig(cfg *vault.Config) {
	v.cfg = cfg
}

// SetToken sets the root token for the provided vault client.
func (v *VaultAPI) SetToken(client *vault.Client, t string) {
	client.SetToken(t)
}

func (v *VaultAPI) Write(client *vault.Client, path string, data map[string]interface{}) (*vault.Secret, error) {
	logical := client.Logical()
	secret, err := logical.Write(path, data)
	return secret, err
}

func (v *VaultAPI) Read(client *vault.Client, path string) (*vault.Secret, error) {
	logical := client.Logical()
	return logical.Read(path)
}

// Delete removes a path from a backend.
func (v *VaultAPI) Delete(client *vault.Client, path string) (*vault.Secret, error) {
	return client.Logical().Delete(path)
}

// Revoke revokes the object represented by the passed-in id.
func (v *VaultAPI) Revoke(client *vault.Client, id string) error {
	return client.Sys().Revoke(id)
}

// VaultAPIConfig contains the applications configuration settings.
type VaultAPIConfig struct {
	ParentToken string // Other tokens will be children of this token.
	Host        string // The hostname or ip address of the vault server.
	Port        string // The port of the vault server.
	Scheme      string // The scheme for vault URL. Should be either http or https.
	CACert      string // The path to the PEM-encoded CA cert file used to verify the Vault server SSL cert.
	ClientCert  string // The path to the client cert used for Vault communication.
	ClientKey   string // The paht to the client key used for Vault communication.
}
