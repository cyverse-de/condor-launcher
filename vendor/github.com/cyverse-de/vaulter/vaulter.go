package vaulter

import (
	"fmt"

	vault "github.com/hashicorp/vault/api"
)

// InitAPI initializes the provided *VaultAPI. This should be called first.
func InitAPI(api *VaultAPI, cfg *VaultAPIConfig, token string) error {
	var err error
	tlsconfig := &vault.TLSConfig{
		CACert:     cfg.CACert,
		ClientCert: cfg.ClientCert,
		ClientKey:  cfg.ClientKey,
	}
	apicfg := api.DefaultConfig()
	apicfg.Address = fmt.Sprintf(
		"%s://%s:%s",
		cfg.Scheme,
		cfg.Host,
		cfg.Port,
	)
	if err = api.ConfigureTLS(apicfg, tlsconfig); err != nil {
		return err
	}
	var client *vault.Client
	if client, err = api.NewClient(apicfg); err != nil {
		return err
	}
	api.SetToken(client, token)
	api.SetClient(client)
	api.SetConfig(apicfg)
	return nil
}
