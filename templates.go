package main

import (
	"bytes"
	"text/template"

	"github.com/pkg/errors"
)

var (
	// IRODSConfigTemplate is the *template.Template for the iRODS config file
	IRODSConfigTemplate *template.Template
)

// IRODSConfigTemplateText is the text of the template for porklock's iRODS
// config file.
const IRODSConfigTemplateText = `porklock.irods-host = {{.IRODSHost}}
porklock.irods-port = {{.IRODSPort}}
porklock.irods-user = {{.IRODSUser}}
porklock.irods-pass = {{.IRODSPass}}
porklock.irods-home = {{.IRODSBase}}
porklock.irods-zone = {{.IRODSZone}}
porklock.irods-resc = {{.IRODSResc}}
`

// IRODSConfig contains all of the values for the IRODS configuration file used
// by the porklock tool out on a HTCondor compute node.
type IRODSConfig struct {
	IRODSHost string
	IRODSPort string
	IRODSUser string
	IRODSPass string
	IRODSZone string
	IRODSBase string
	IRODSResc string
}

// GenerateFile applies the data to the given template and returns a *bytes.Buffer
// containing the result.
func GenerateFile(t *template.Template, data interface{}) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	err := t.Execute(&buffer, data)
	if err != nil {
		return &buffer, errors.Wrapf(err, "failed to apply data to the %s template", t.Name())
	}
	return &buffer, err
}
