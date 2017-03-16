package main

import (
	"bytes"
	"text/template"

	"github.com/cyverse-de/model"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
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
    filter_files: "{{.GetString "condor.filter_files"}}"`

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

type irodsconfig struct {
	IRODSHost string
	IRODSPort string
	IRODSUser string
	IRODSPass string
	IRODSZone string
	IRODSBase string
	IRODSResc string
}

// GenerateIRODSConfig returns the contents of the irods-config file as a string.
func GenerateIRODSConfig(t *template.Template, cfg *viper.Viper) (*bytes.Buffer, error) {
	c := &irodsconfig{
		IRODSHost: cfg.GetString("irods.host"),
		IRODSPort: cfg.GetString("irods.port"),
		IRODSUser: cfg.GetString("irods.user"),
		IRODSPass: cfg.GetString("irods.pass"),
		IRODSBase: cfg.GetString("irods.base"),
		IRODSResc: cfg.GetString("irods.resc"),
		IRODSZone: cfg.GetString("irods.zone"),
	}
	var buffer bytes.Buffer
	err := t.Execute(&buffer, c)
	if err != nil {
		return &buffer, errors.Wrap(err, "failed to apply data to the irods config template")
	}
	return &buffer, err
}

// GenerateCondorSubmit returns a string (or error) containing the contents
// of what should go into an HTCondor submission file.
func GenerateCondorSubmit(t *template.Template, submission *model.Job) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	err := t.Execute(&buffer, submission)
	if err != nil {
		return &buffer, errors.Wrap(err, "failed to apply data to the submission template")
	}
	return &buffer, nil
}

// GenerateJobConfig creates a string containing the config that gets passed
// into the job.
func GenerateJobConfig(t *template.Template, cfg *viper.Viper) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	err := t.Execute(&buffer, cfg)
	if err != nil {
		return &buffer, errors.Wrap(err, "failed to apply data to the job config template")
	}
	return &buffer, nil
}
