package vaulter

import (
	"fmt"
	"strconv"

	vault "github.com/hashicorp/vault/api"
)

// RoleConfig contains the settings applied to a new role.
type RoleConfig struct {
	AllowedDomains  string
	AllowSubdomains bool
	KeyBits         int
	MaxTTL          string
	AllowAnyName    bool
}

// CreateRole creates a new role.
func CreateRole(r MountReaderWriter, mountPath, roleName string, c *RoleConfig) (*vault.Secret, error) {
	client := r.Client()
	writePath := fmt.Sprintf("%s/roles/%s", mountPath, roleName)
	data := map[string]interface{}{
		"allowed_domains":  c.AllowedDomains,
		"allow_subdomains": strconv.FormatBool(c.AllowSubdomains),
		"key_bits":         c.KeyBits,
		"allow_any_name":   strconv.FormatBool(c.AllowAnyName),
	}
	return r.Write(client, writePath, data)
}

// HasRole returns true if the passed in role exists and has the same settings.
func HasRole(r MountReaderWriter, mountPath, roleName, domains string, subdomains bool) (bool, error) {
	client := r.Client()
	readPath := fmt.Sprintf("%s/roles/%s", mountPath, roleName)
	secret, err := r.Read(client, readPath)
	if err != nil {
		return false, err
	}
	if secret == nil {
		return false, nil
	}
	if secret.Data == nil {
		return false, nil
	}
	v, ok := secret.Data["allowed_domains"]
	if !ok {
		return false, nil
	}
	if v != domains {
		return false, nil
	}
	v, ok = secret.Data["allow_subdomains"]
	if !ok {
		return false, nil
	}
	if v != subdomains {
		fmt.Printf("v: %s\tsubdomains: %s\n", v, strconv.FormatBool(subdomains))
		return false, nil
	}
	return true, nil
}
