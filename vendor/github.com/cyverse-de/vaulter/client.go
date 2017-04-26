package vaulter

import vault "github.com/hashicorp/vault/api"

// TokenSetter is an interface for objects that can set the root token for Vault
// clients.
type TokenSetter interface {
	SetToken(c *vault.Client, t string)
}

// ConfigGetter is an interface for objects that need access to the
// *vault.Config.
type ConfigGetter interface {
	GetConfig() *vault.Config
}

// ConfigSetter is an interface for objects that need to set the *vault.Config
// for an underlying client.
type ConfigSetter interface {
	SetConfig(c *vault.Config)
}

// Configurer is an interface for objects that can configure a Vault client.
type Configurer interface {
	DefaultConfigurer
	ConfigureTLS(t *vault.TLSConfig) error
}

// DefaultConfigurer is an interface for objects that want a default Vault
// client configuration instance.
type DefaultConfigurer interface {
	DefaultConfig() *vault.Config
}

// ClientCreator is an interface for objects that can create new Vault clients.
type ClientCreator interface {
	NewClient(c *vault.Config) (*vault.Client, error)
}

// ClientGetter is an interface for objects that need access to the Vault
// client.
type ClientGetter interface {
	Client() *vault.Client
}

// ClientSetter is an interface for objects that need to set their internal
// client value.
type ClientSetter interface {
	SetClient(c *vault.Client)
}

// ClientWriter defines the interface for writing data to a mount after
// creating a new Vault API client.
type ClientWriter interface {
	ClientCreator
	ConfigGetter
	DefaultConfigurer
	TokenSetter
	MountWriter
}

// ClientReader defines the interface for reading data from a mount after
// creating a new Vault API client.
type ClientReader interface {
	ClientCreator
	ConfigGetter
	TokenSetter
	MountReader
}
