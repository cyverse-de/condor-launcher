package vaulter

import (
	"fmt"
	"strings"

	vault "github.com/hashicorp/vault/api"
)

// Revoker is an interface for objects that can be used to revoke a certificate.
type Revoker interface {
	Revoke(c *vault.Client, id string) error
}

// PKIChecker defines the interface for checking to see if root PKI cert is
// configured.
type PKIChecker interface {
	ClientGetter
	MountWriter // this is not a mistake.
}

// HasRootCert returns true if a cert for the provided role and common-name
// already exists. The current process is a hack. We attempt to generate a cert,
// if the attempt succeeds then the root cert exists.
func HasRootCert(m PKIChecker, mount, role, commonName string) (bool, error) {
	var (
		client *vault.Client
		err    error
	)
	client = m.Client()
	writePath := fmt.Sprintf("%s/issue/%s", mount, role)
	_, err = m.Write(client, writePath, map[string]interface{}{
		"common_name": commonName,
	})
	if err != nil {
		if strings.HasSuffix(err.Error(), "backend must be configured with a CA certificate/key") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CSRConfig contains configuration settings for generating a certificate
// signing request.
type CSRConfig struct {
	CommonName        string
	TTL               string
	KeyBits           int
	ExcludeCNFromSans bool // disables adding the common name to the list of subject alternative names
}

// ImportCert sets the signed cert for the backend mounted at the given path
// inside Vault.
func ImportCert(m MountReaderWriter, mountPath, certContents string) (*vault.Secret, error) {
	client := m.Client()
	path := fmt.Sprintf("%s/intermediate/set-signed", mountPath)
	data := map[string]interface{}{
		"certificate": certContents,
	}
	return m.Write(client, path, data)
}

// CSR generates a certificate signing request using the backend mounted at the
// provided directory.
func CSR(m MountReaderWriter, mountPath string, c *CSRConfig) (*vault.Secret, error) {
	var client *vault.Client
	client = m.Client()
	path := fmt.Sprintf("%s/intermediate/generate/internal", mountPath)
	data := map[string]interface{}{
		"common_name":          c.CommonName,
		"ttl":                  c.TTL,
		"key_bits":             c.KeyBits,
		"exclude_cn_from_sans": c.ExcludeCNFromSans,
	}
	return m.Write(client, path, data)
}

// RootCACertConfig contains the settings for the root CA cert.
type RootCACertConfig struct {
	CommonName        string
	TTL               string
	KeyBits           int
	ExcludeCNFromSans bool
}

// RootCACert generates the root CA cert and key using the backend mounted at
// the provided directory.
func RootCACert(m MountReaderWriter, mountPath string, c *RootCACertConfig) (*vault.Secret, error) {
	var client *vault.Client
	client = m.Client()
	path := fmt.Sprintf("%s/root/generate/internal", mountPath)
	data := map[string]interface{}{
		"common_name":          c.CommonName,
		"ttl":                  c.TTL,
		"key_bits":             c.KeyBits,
		"exclude_cn_from_sans": c.ExcludeCNFromSans,
	}
	return m.Write(client, path, data)
}

// CSRSigningConfig contains the configuration settings for signing a CSR.
type CSRSigningConfig struct {
	CommonName string
	TTL        string
}

// SignCSR signs the provided CSR by passing it to the /root/sign-intermediate
// path in the given backend.
func SignCSR(m MountReaderWriter, rootPath string, csr string, c *CSRSigningConfig) (*vault.Secret, error) {
	client := m.Client()
	path := fmt.Sprintf("%s/root/sign-intermediate", rootPath)
	data := map[string]interface{}{
		"common_name": c.CommonName,
		"ttl":         c.TTL,
		"csr":         csr,
	}
	return m.Write(client, path, data)
}

// ConfigCAAccess sets the issuing_certificates and crl_distribution_points URLs
// for the backend mounted at the given path.
func ConfigCAAccess(m MountReaderWriter, scheme, hostPort, mountPath string) (*vault.Secret, error) {
	var client *vault.Client
	client = m.Client()
	path := fmt.Sprintf("%s/config/urls", mountPath)
	data := map[string]interface{}{
		"issuing_certificates":    fmt.Sprintf("%s://%s/v1/%s/ca", scheme, hostPort, mountPath),
		"crl_distribution_points": fmt.Sprintf("%s://%s/v1/%s/crl", scheme, hostPort, mountPath),
	}
	return m.Write(client, path, data)
}

// IssueCertConfig contains the settings needed for issuing a cert.
type IssueCertConfig struct {
	CommonName        string
	AltNames          string // csv of requested subject alternative names
	IPSans            string // csv of ip subject alternative names
	TTL               string
	Format            string // See the /pki/issue docs on https://www.vaultproject.io/docs/secrets/pki/ for valid values.
	ExcludeCNFromSans bool   // exclude common name from subject alternative names
}

// IssueCert issues a cert with the given backend using the given role name.
func IssueCert(m MountReaderWriter, mountPath, roleName string, c *IssueCertConfig) (*vault.Secret, error) {
	client := m.Client()
	path := fmt.Sprintf("%s/issue/%s", mountPath, roleName)
	data := map[string]interface{}{
		"common_name":          c.CommonName,
		"alt_names":            c.AltNames,
		"ip_sans":              c.IPSans,
		"ttl":                  c.TTL,
		"format":               c.Format,
		"exclude_cn_from_sans": c.ExcludeCNFromSans,
	}
	return m.Write(client, path, data)
}
