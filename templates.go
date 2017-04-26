package main

import (
	"bytes"
	"text/template"

	"github.com/pkg/errors"
)

var (
	// SubmissionTemplate is a *template.Template for the HTCondor submission file
	SubmissionTemplate *template.Template

	// JobConfigTemplate is the *template.Template for the job definition JSON
	JobConfigTemplate *template.Template

	// IRODSConfigTemplate is the *template.Template for the iRODS config file
	IRODSConfigTemplate *template.Template
)

// SubmissionTemplateText is the text of the template for the HTCondor
// submission file.
const SubmissionTemplateText = `universe = vanilla
executable = /usr/local/bin/road-runner
rank = mips{{ if .UsesVolumes }}
requirements = (HAS_HOST_MOUNTS == True){{ end }}
arguments = --config config --job job
output = script-output.log
error = script-error.log
log = condor.log{{if .Group}}
accounting_group = {{.Group}}
accounting_group_user = {{.Submitter}}{{end}}
request_disk = {{.RequestDisk}}
+IpcUuid = "{{.InvocationID}}"
+IpcJobId = "generated_script"
+IpcUsername = "{{.Submitter}}"
+IpcUserGroups = {{.FormatUserGroups}}
concurrency_limits = {{.UserIDForSubmission}}
{{with $x := index .Steps 0}}+IpcExe = "{{$x.Component.Name}}"{{end}}
{{with $x := index .Steps 0}}+IpcExePath = "{{$x.Component.Location}}"{{end}}
should_transfer_files = YES
transfer_input_files = irods-config,iplant.cmd,config,job
transfer_output_files = workingvolume/logs/logs-stdout-output,workingvolume/logs/logs-stderr-output
when_to_transfer_output = ON_EXIT_OR_EVICT
notification = NEVER
queue
`

// JobConfigTemplateText is the text of the template for the HTCondor submission
// file.
const JobConfigTemplateText = `amqp:
    uri: {{.GetString "amqp.uri"}}
    exchange:
        name: {{.GetString "amqp.exchange.name"}}
        type: {{.GetString "amqp.exchange.type"}}
irods:
    base: "{{.GetString "irods.base"}}"
porklock:
    image: "{{.GetString "porklock.image"}}"
    tag: "{{.GetString "porklock.tag"}}"
condor:
    filter_files: "{{.GetString "condor.filter_files"}}"
vault:
    token: "{{.GetString "vault.child_token.token"}}"`

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
