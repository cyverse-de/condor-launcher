package vaulter

import (
	"errors"
	"strings"

	vault "github.com/hashicorp/vault/api"
)

// Mounter is an interface for objects that can mount Vault backends.
type Mounter interface {
	Mount(path string, m *vault.MountInput) error
}

// MountLister is an interface for objects that can list mounted Vault backends.
type MountLister interface {
	ListMounts() (map[string]*vault.MountOutput, error)
}

// MountConfigGetter is an interface for objects that can get the configuration
// for a mount in Vault.
type MountConfigGetter interface {
	MountConfig(path string) (*vault.MountConfigOutput, error)
}

// MountTuner is an interface for objects that need to configure a mount in
// Vault.
type MountTuner interface {
	TuneMount(path string, input vault.MountConfigInput) error
}

// MountWriter is an interface for objects that can write to a path in a Vault
// backend.
type MountWriter interface {
	Write(c *vault.Client, path string, data map[string]interface{}) (*vault.Secret, error)
}

// MountReader is an interface for objects that can read data from a path in a
// Vault backend.
type MountReader interface {
	Read(c *vault.Client, path string) (*vault.Secret, error)
}

// PathDeleter is an interface for deleting information from a mount, not for
// deleting the mount itself.
type PathDeleter interface {
	Delete(c *vault.Client, path string) (*vault.Secret, error)
}

// Unmounter is an interface for objects that can unmount a Vault
// backend.
type Unmounter interface {
	Unmount(path string) error
}

// MountReaderWriter defines an interface for doing role related operations.
type MountReaderWriter interface {
	ClientGetter
	MountWriter
	MountReader
}

// MountDeleter defines and interface for deleting content from a path in a
// mounted backend.
type MountDeleter interface {
	ClientGetter
	PathDeleter
}

// MountConfiguration is a flattened representation of the configs that the Vault API
// supports for the backend mounts.
type MountConfiguration struct {
	Type            string
	Description     string
	DefaultLeaseTTL string
	MaxLeaseTTL     string
}

// Mount mounts a vault backend with the provided
func Mount(m Mounter, path string, c *MountConfiguration) error {
	return m.Mount(path, &vault.MountInput{
		Type:        c.Type,
		Description: c.Description,
		Config: vault.MountConfigInput{
			DefaultLeaseTTL: c.DefaultLeaseTTL,
			MaxLeaseTTL:     c.MaxLeaseTTL,
		},
	})
}

// Unmount unmounts a vault backend with the provided path.
func Unmount(u Unmounter, path string) error {
	return u.Unmount(path)
}

// MountConfig returns the config for the passed in mount rooted at the given
// path.
func MountConfig(m MountConfigGetter, path string) (*vault.MountConfigOutput, error) {
	return m.MountConfig(path)
}

// IsMounted returns true if the given path is mounted as a backend in Vault.
func IsMounted(l MountLister, path string) (bool, error) {
	var (
		hasPath bool
		err     error
	)
	mounts, err := l.ListMounts()
	if err != nil {
		return false, err
	}
	for m := range mounts {
		if strings.TrimSuffix(m, "/") == path {
			hasPath = true
		}
	}
	return hasPath, nil
}

// WriteMount writes data to a path in a backend using a newly created
// client whose token is set to the one provided.
func WriteMount(cw ClientWriter, path, token string, data map[string]interface{}) error {
	var (
		client *vault.Client
		err    error
	)
	defcfg := cw.DefaultConfig()
	newcfg := cw.GetConfig()
	defcfg.Address = newcfg.Address
	defcfg.MaxRetries = newcfg.MaxRetries
	if client, err = cw.NewClient(defcfg); err != nil {
		return err
	}
	cw.SetToken(client, token)
	_, err = cw.Write(client, path, data)
	if err != nil {
		return err
	}
	return nil
}

// ReadMount reads data from a path in a mount using a newly created client
// whose token is set to the one provided.
func ReadMount(cr ClientReader, path, token string) (map[string]interface{}, error) {
	var (
		client *vault.Client
		err    error
	)
	if client, err = cr.NewClient(cr.GetConfig()); err != nil {
		return nil, err
	}
	cr.SetToken(client, token)
	secret, err := cr.Read(client, path)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, errors.New("secret is nil")
	}
	if secret.Data == nil {
		return nil, errors.New("data is nil")
	}
	return secret.Data, nil
}

// Delete deletes data from the path in the mount. Does not delete a mount.
// You unmount a mount, you don't delete one.
func Delete(md MountDeleter, path string) (*vault.Secret, error) {
	return md.Delete(md.Client(), path)
}
