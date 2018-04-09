package jobs

import (
	"text/template"
	"log"
	"github.com/pkg/errors"
	"gopkg.in/cyverse-de/model.v2"
)

type OtherTemplateFields struct {
	TicketPathListHeader string `json:"ticket_path_list_header"`
}

type TemplatesModel struct {
	*model.Job
	OtherTemplateFields
}

var (
	// condorSubmissionTemplate is a *template.Template for the HTCondor submission file
	condorSubmissionTemplate *template.Template

	// condorJobConfigTemplate is the *template.Template for the job definition JSON
	condorJobConfigTemplate *template.Template

	// The *template.Template for a list of input files with iRODS download tickets.
	inputTicketListTemplate *template.Template

	// The *template.Template for the iRODS output dest with ticket.
	outputTicketListTemplate *template.Template
)

// SubmissionTemplateText is the text of the template for the HTCondor
// submission file.
const condorSubmissionTemplateText =
`universe = vanilla
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
transfer_input_files = iplant.cmd,config,job{{if .OutputTicketFile}},{{.OutputTicketFile}}{{end}}{{if .InputTicketsFile}},{{.InputTicketsFile}}{{end}}
transfer_output_files = workingvolume/logs/logs-stdout-output,workingvolume/logs/logs-stderr-output
when_to_transfer_output = ON_EXIT_OR_EVICT
notification = NEVER
queue
`

// JobConfigTemplateText is the text of the template for the HTCondor submission
// file.
const condorJobConfigTemplateText =
`amqp:
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
`

// The text of the template for a list of input files with iRODS download tickets.
const inputTicketListTemplateText =
`{{.TicketPathListHeader}}
{{range .FilterInputsWithTickets -}}
{{.Ticket}},{{.IRODSPath}}
{{end}}`

// The text of the template for the iRODS output dest with ticket.
const outputTicketListTemplateText =
`{{.TicketPathListHeader}}
{{.OutputDirTicket}},{{.OutputDir}}
`

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

	inputTicketListTemplate, err = template.New("input_tickets").Parse(inputTicketListTemplateText)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse input tickets template text"))
	}

	outputTicketListTemplate, err = template.New("output_ticket").Parse(outputTicketListTemplateText)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse output ticket template text"))
	}
}
