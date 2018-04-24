package jobs

import (
	"log"
	"text/template"

	"github.com/pkg/errors"
)

var (
	// condorSubmissionTemplate is a *template.Template for the HTCondor submission file
	condorSubmissionTemplate *template.Template

	// condorJobConfigTemplate is the *template.Template for the job definition JSON
	condorJobConfigTemplate *template.Template

	// interappsSubmissionTemplate is a *template.Template fo the HTCondor
	// submission files that define an interactive app job.
	interappsSubmissionTemplate *template.Template

	// interappsJobConfigTemplate ia *template.Template for the job config needed
	// for interactive applications.
	interappsJobConfigTemplate *template.Template
)

// SubmissionTemplateText is the text of the template for the HTCondor
// submission file.
const condorSubmissionTemplateText = `universe = vanilla
executable = /usr/local/bin/road-runner
rank = mips{{ if .UsesVolumes }}
requirements = (HAS_HOST_MOUNTS == True){{ end }}
arguments = --config config --job job
output = script-output.log
error = script-error.log
log = condor.log
accounting_group = {{if .Group}}{{.Group}}{{else}}de{{end}}
accounting_group_user = {{.Submitter}}
request_disk = {{.RequestDisk}}
+IpcUuid = "{{.InvocationID}}"
+IpcJobId = "generated_script"
+IpcUsername = "{{.Submitter}}"
+IpcUserGroups = {{.FormatUserGroups}}
concurrency_limits = {{.UserIDForSubmission}}
{{with $x := index .Steps 0}}+IpcExe = "{{$x.Component.Name}}"{{end}}
{{with $x := index .Steps 0}}+IpcExePath = "{{$x.Component.Location}}"{{end}}
should_transfer_files = YES
transfer_input_files = iplant.cmd,config,job
transfer_output_files = workingvolume/logs/logs-stdout-output,workingvolume/logs/logs-stderr-output
when_to_transfer_output = ON_EXIT_OR_EVICT
notification = NEVER
queue
`

// JobConfigTemplateText is the text of the template for the HTCondor submission
// file.
const condorJobConfigTemplateText = `amqp:
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
    token: "{{.GetString "vault.child_token.token"}}"
    url: "{{.GetString "vault.url"}}"`

const interappsSubmissionTemplateText = `universe = vanilla
executable = /usr/local/bin/interapps-runner
rank = mips{{ if .UsesVolumes }}
requirements = (HAS_HOST_MOUNTS == True){{ end }}
arguments = --config config --job job
output = script-output.log
error = script-error.log
log = condor.log
accounting_group = {{if .Group}}{{.Group}}{{else}}de{{end}}
accounting_group_user = {{.Submitter}}
request_disk = {{.RequestDisk}}
+IpcUuid = "{{.InvocationID}}"
+IpcJobId = "generated_script"
+IpcUsername = "{{.Submitter}}"
+IpcUserGroups = {{.FormatUserGroups}}
concurrency_limits = {{.UserIDForSubmission}}
{{with $x := index .Steps 0}}+IpcExe = "{{$x.Component.Name}}"{{end}}
{{with $x := index .Steps 0}}+IpcExePath = "{{$x.Component.Location}}"{{end}}
should_transfer_files = YES
transfer_input_files = iplant.cmd,config,job
transfer_output_files = workingvolume/logs/logs-stdout-output,workingvolume/logs/logs-stderr-output
when_to_transfer_output = ON_EXIT_OR_EVICT
notification = NEVER
queue`

const interappsJobConfigTemplateText = `amqp:
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
    token: "{{.GetString "vault.child_token.token"}}"
    url: "{{.GetString "vault.url"}}"
k8s:
    app-exposer:
        base: {{.GetString "k8s.app-exposer.base"}}`

func init() {
	var err error
	condorSubmissionTemplate, err = template.New("condor_submit").Parse(condorSubmissionTemplateText)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse submission template text"))
	}
	condorJobConfigTemplate, err = template.New("job_config").Parse(condorJobConfigTemplateText)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse job config template text"))
	}
	interappsSubmissionTemplate, err = template.New("interapps_condor_submit").Parse(interappsSubmissionTemplateText)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse interapps submission template text"))
	}
	interappsJobConfigTemplate, err = template.New("interapps_job_config").Parse(interappsJobConfigTemplateText)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse interapps job config template text"))
	}
}
