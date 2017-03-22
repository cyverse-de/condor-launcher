package main

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/cyverse-de/vaulter"
	vault "github.com/hashicorp/vault/api"
)

// VaultOperator defines the operations that can be performed on a Vault.
type VaultOperator interface {
	MountCubbyhole(mountPoint string) error
	ChildToken(numUses int) (string, error)
	StoreConfig(token, mountPoint, jobID string, config []byte) error
}

// Vaulter is a type of VaultOperator that can interact with an actual Vault
// service.
type Vaulter struct {
	api *vaulter.VaultAPI
}

// VaultInit initializes a *Vaulter.
func VaultInit(token, apiurl string) (*Vaulter, error) {
	urlParts, err := url.Parse(apiurl)
	if err != nil {
		return nil, err
	}
	cfg := &vaulter.VaultAPIConfig{
		ParentToken: token,
		Host:        urlParts.Hostname(),
		Port:        urlParts.Port(),
		Scheme:      urlParts.Scheme,
	}
	api := &vaulter.VaultAPI{}
	if err = vaulter.InitAPI(api, cfg, cfg.ParentToken); err != nil {
		return nil, err
	}
	return &Vaulter{
		api: api,
	}, nil
}

// MountCubbyhole will mount a cubbyhole backend in Vault at the provided
// mount point.
func (v *Vaulter) MountCubbyhole(mountPoint string) error {
	hasMount, err := vaulter.IsMounted(v.api, mountPoint)
	if err != nil {
		return err
	}
	if !hasMount {
		if err = vaulter.Mount(v.api, mountPoint, &vaulter.MountConfiguration{
			Type:        "cubbyhole",
			Description: "A cubbyhole for the iRODS configs used in jobs",
		}); err != nil {
			return err
		}
	}
	return nil
}

// ChildToken creates a child token of the configured root token with a limited
// number of uses.
func (v *Vaulter) ChildToken(numUses int) (string, error) {
	opts := &vault.TokenCreateRequest{
		NumUses: numUses,
	}
	ta := v.api.Token()
	secret, err := v.api.CreateToken(ta, opts)
	if err != nil {
		return "", nil
	}
	if secret.Auth == nil {
		return "", errors.New("auth field was nil")
	}
	if secret.Auth.ClientToken == "" {
		return "", errors.New("client token was empty")
	}
	return secret.Auth.ClientToken, nil
}

// StoreConfig will store the provided config in the mount indicated by
// mountPoint.
func (v *Vaulter) StoreConfig(token, mountPoint, jobID string, config []byte) error {
	data := map[string]interface{}{
		"config": config,
	}
	cubbypath := fmt.Sprintf("%s/%s", mountPoint, jobID)
	return vaulter.WriteMount(v.api, cubbypath, token, data)
}
