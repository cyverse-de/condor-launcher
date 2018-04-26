package vaulter

import (
	"errors"
	"testing"

	vault "github.com/hashicorp/vault/api"
)

type StubMountReaderWriter struct {
	path       string
	data       map[string]interface{}
	writeError bool
	readError  bool
}

func (r *StubMountReaderWriter) Client() *vault.Client {
	return &vault.Client{}
}

func (r *StubMountReaderWriter) Write(client *vault.Client, path string, data map[string]interface{}) (*vault.Secret, error) {
	r.data = data
	secret := &vault.Secret{}
	r.path = path
	if r.writeError {
		return nil, errors.New("write error")
	}
	return secret, nil
}

func (r *StubMountReaderWriter) Read(client *vault.Client, path string) (*vault.Secret, error) {
	if r.readError {
		return nil, errors.New("read error")
	}
	r.path = path
	retval := &vault.Secret{
		Data: map[string]interface{}{
			"allowed_domains":  "foo.com",
			"allow_subdomains": "true",
		},
	}
	return retval, nil
}

type StubPKIChecker struct {
	cfg           *vault.Config
	token         string
	path          string
	data          map[string]interface{}
	clientError   bool
	writeError    bool
	notFoundError bool
}

func (w *StubPKIChecker) Client() *vault.Client {
	return &vault.Client{}
}

func (w *StubPKIChecker) Write(client *vault.Client, token string, data map[string]interface{}) (*vault.Secret, error) {
	w.data = data
	secret := &vault.Secret{}
	if w.notFoundError {
		return nil, errors.New("backend must be configured with a CA certificate/key")
	}
	if w.writeError {
		return secret, errors.New("write error")
	}
	return secret, nil
}

func TestHasRootCert(t *testing.T) {
	cw := &StubPKIChecker{notFoundError: true}
	hasCert, err := HasRootCert(cw, "pki", "example-dot-com", "test.example.com")
	if err != nil {
		t.Error(err)
	}
	if hasCert {
		t.Error("cert was found when it should be missing")
	}

	cw = &StubPKIChecker{writeError: true}
	hasCert, err = HasRootCert(cw, "pki", "example-dot-com", "test.example.com")
	if err == nil {
		t.Error("err was nil when it should have been set")
	}
	if hasCert {
		t.Error("cert was found when it should be missing")
	}

	cw = &StubPKIChecker{}
	hasCert, err = HasRootCert(cw, "pki", "example-dot-com", "test.example.com")
	if err != nil {
		t.Error(err)
	}
	if !hasCert {
		t.Error("cert was not found when it should be present")
	}
}

func TestImportCert(t *testing.T) {
	rw := &StubMountReaderWriter{}
	s, err := ImportCert(rw, "test", "foo")
	if err != nil {
		t.Error(err)
	}
	if s == nil {
		t.Error("ImportSecret() returned a nil secret")
	}
	if rw.data["certificate"] != "foo" {
		t.Errorf("certificate was set to %s instead of 'foo'", rw.data["certificate"])
	}
	expected := "test/intermediate/set-signed"
	if rw.path != expected {
		t.Errorf("path is '%s' instead of '%s'", rw.path, expected)
	}
}

func TestCSR(t *testing.T) {
	rw := &StubMountReaderWriter{}
	cfg := &CSRConfig{
		CommonName:        "common.name",
		TTL:               "forever",
		KeyBits:           4096,
		ExcludeCNFromSans: true,
	}
	s, err := CSR(rw, "test", cfg)
	if err != nil {
		t.Error(err)
	}
	if s == nil {
		t.Error("s is nil")
	}
	expected := "common.name"
	if rw.data["common_name"] != expected {
		t.Errorf("common_name was %s instead of %s", rw.data["common_name"], expected)
	}
	expected = "forever"
	if rw.data["ttl"] != expected {
		t.Errorf("ttl was %s instead of %s", rw.data["ttl"], expected)
	}
	expectedbits := 4096
	if rw.data["key_bits"] != expectedbits {
		t.Errorf("key_bits was %d instead of %d", rw.data["key_bits"], expectedbits)
	}
	if !rw.data["exclude_cn_from_sans"].(bool) {
		t.Errorf("exclude_cn_from_sans was false instead of true")
	}
}

func TestRootCACert(t *testing.T) {
	rw := &StubMountReaderWriter{}
	cfg := &RootCACertConfig{
		CommonName:        "common.name",
		TTL:               "forever",
		KeyBits:           4096,
		ExcludeCNFromSans: true,
	}
	s, err := RootCACert(rw, "test", cfg)
	if err != nil {
		t.Error(err)
	}
	if s == nil {
		t.Error("s is nil")
	}
	expected := "common.name"
	if rw.data["common_name"] != expected {
		t.Errorf("common_name was %s instead of %s", rw.data["common_name"], expected)
	}
	expected = "forever"
	if rw.data["ttl"] != expected {
		t.Errorf("ttl was %s instead of %s", rw.data["ttl"], expected)
	}
	expectedbits := 4096
	if rw.data["key_bits"] != expectedbits {
		t.Errorf("key_bits was %d instead of %d", rw.data["key_bits"], expectedbits)
	}
	if !rw.data["exclude_cn_from_sans"].(bool) {
		t.Errorf("exclude_cn_from_sans was false instead of true")
	}
}

func TestSignCSR(t *testing.T) {
	rw := &StubMountReaderWriter{}
	cfg := &CSRSigningConfig{
		CommonName: "common.name",
		TTL:        "forever",
	}
	s, err := SignCSR(rw, "test", "test-csr", cfg)
	if err != nil {
		t.Error(err)
	}
	if s == nil {
		t.Error("s is nil")
	}
	expected := "common.name"
	if rw.data["common_name"] != expected {
		t.Errorf("common_name was %s instead of %s", rw.data["common_name"], expected)
	}
	expected = "forever"
	if rw.data["ttl"] != expected {
		t.Errorf("ttl was %s instead of %s", rw.data["ttl"], expected)
	}
	expected = "test-csr"
	if rw.data["csr"] != expected {
		t.Errorf("csr was %s instead of %s", rw.data["csr"], expected)
	}
}

func TestConfigCAAccess(t *testing.T) {
	rw := &StubMountReaderWriter{}
	s, err := ConfigCAAccess(rw, "https", "test:12345", "test")
	if err != nil {
		t.Error(err)
	}
	if s == nil {
		t.Error("s was nil")
	}
	expected := "https://test:12345/v1/test/ca"
	actual := rw.data["issuing_certificates"]
	if actual != expected {
		t.Errorf("issuing_certificates was '%s' instead of '%s'", actual, expected)
	}
	expected = "https://test:12345/v1/test/crl"
	actual = rw.data["crl_distribution_points"]
	if actual != expected {
		t.Errorf("crl_distribution_points was '%s' instead of '%s'", actual, expected)
	}
}

func TestIssueCert(t *testing.T) {
	rw := &StubMountReaderWriter{}
	cfg := &IssueCertConfig{
		CommonName:        "common.name",
		AltNames:          "alt-names",
		IPSans:            "ip-sans",
		TTL:               "forever",
		Format:            "format",
		ExcludeCNFromSans: true,
	}
	s, err := IssueCert(rw, "test-mount", "test-role", cfg)
	if err != nil {
		t.Error(err)
	}
	if s == nil {
		t.Error("s was nil")
	}
	fields := map[string]string{
		"common_name": "common.name",
		"alt_names":   "alt-names",
		"ip_sans":     "ip-sans",
		"ttl":         "forever",
		"format":      "format",
	}

	for k, expected := range fields {
		actual := rw.data[k]
		if actual != expected {
			t.Errorf("rw.data[\"%s\"] => %s, expected => %s", k, actual, expected)
		}
	}
	actualb := rw.data["exclude_cn_from_sans"].(bool)
	if !actualb {
		t.Error("exclude_cn_from_sans was false")
	}
}
