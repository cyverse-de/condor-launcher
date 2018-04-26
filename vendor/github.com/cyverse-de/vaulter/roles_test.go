package vaulter

import (
	"errors"
	"testing"

	vault "github.com/hashicorp/vault/api"
)

type StubRoller struct {
	path       string
	data       map[string]interface{}
	writeError bool
	readError  bool
}

func (r *StubRoller) Client() *vault.Client {
	return &vault.Client{}
}

func (r *StubRoller) Write(client *vault.Client, token string, data map[string]interface{}) (*vault.Secret, error) {
	r.data = data
	secret := &vault.Secret{}
	if r.writeError {
		return nil, errors.New("write error")
	}
	return secret, nil
}

func (r *StubRoller) Read(client *vault.Client, path string) (*vault.Secret, error) {
	if r.readError {
		return nil, errors.New("read error")
	}
	r.path = path
	retval := &vault.Secret{
		Data: map[string]interface{}{
			"allowed_domains":  "foo.com",
			"allow_subdomains": true,
		},
	}
	return retval, nil
}

func TestCreateRole(t *testing.T) {
	sr := &StubRoller{}
	rc := &RoleConfig{
		AllowedDomains:  "foo.com",
		AllowSubdomains: true,
	}
	secret, err := CreateRole(sr, "pki", "foo", rc)
	if err != nil {
		t.Error(err)
	}
	if secret == nil {
		t.Error("secret is nil")
	}

	sr = &StubRoller{writeError: true}
	secret, err = CreateRole(sr, "pki", "foo", rc)
	if err == nil {
		t.Error(err)
	}
	if secret != nil {
		t.Error("secret is nil")
	}
}

func TestHasRole(t *testing.T) {
	sr := &StubRoller{}
	hasRole, err := HasRole(sr, "pki", "foo", "foo.com", true)
	if err != nil {
		t.Error(err)
	}
	if !hasRole {
		t.Error("hasRole was false")
	}
}
