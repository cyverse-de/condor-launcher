package jobs

import (
	"log"
	"text/template"

	"github.com/pkg/errors"
)

var (
	// osgSubmissionTemplate is a *template.Template for the OSG submission file
	osgSubmissionTemplate *template.Template
)

// SubmissionTemplateText is the text of the template for the OSG
// submission file.
const osgSubmissionTemplateText = `universe = vanilla
executable = wrapper
requirements = HAS_SINGULARITY == TRUE

output = script-output.log
error = script-error.log
log = condor.log

{{with $x := index .Steps 0}}+SingularityImage = "{{$x.Component.Container.Image.OSGImagePath}}"{{end}}
+SingularityBindCVMFS = True

+IpcUuid = "{{.InvocationID}}"
+IpcJobId = "generated_script"
+IpcUsername = "{{.Submitter}}"

should_transfer_files = YES
transfer_executable = False
transfer_input_files = iplant.cmd{{if .ConfigFile}},{{.ConfigFile}}{{end}}{{if .OutputTicketFile}},{{.OutputTicketFile}}{{end}}{{if .InputTicketsFile}},{{.InputTicketsFile}}{{end}}
when_to_transfer_output = NEVER
notification = NEVER

queue
`

func init() {
	var err error
	osgSubmissionTemplate, err = template.New("osg_submit").Parse(osgSubmissionTemplateText)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse submission template text"))
	}
}
