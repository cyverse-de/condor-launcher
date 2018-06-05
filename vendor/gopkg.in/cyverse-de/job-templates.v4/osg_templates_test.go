package jobs

import (
	"testing"
)

func TestGenerateOSGSubmit(t *testing.T) {
	cfg := InitConfig(t)
	s := InitTestsFromFile(t, cfg, "osg_submission.json")
	actual, err := generateFileContents(osgSubmissionTemplate, s)
	if err != nil {
		t.Error(err)
	}
	expected := `universe = vanilla
executable = /usr/bin/wrapper
requirements = HAS_SINGULARITY == TRUE

output = script-output.log
error = script-error.log
log = condor.log

+SingularityImage = "/cvmfs/singularity.opensciencegrid.org/discoenv/osg-word-count"
+SingularityBindCVMFS = True

+IpcUuid = "2256dd6d-d984-4d3a-ad71-ab1ff341f636"
+IpcJobId = "generated_script"
+IpcUsername = "sarahr"
+ProjectName = "cyverse"

should_transfer_files = YES
transfer_executable = False
transfer_input_files = iplant.cmd,config.json,output_ticket.list,input_ticket.list
when_to_transfer_output = ON_EXIT_OR_EVICT
notification = NEVER

queue
`
	if actual.String() != expected {
		t.Errorf("GenerateCondorSubmit() returned:\n\n%s\n\ninstead of:\n\n%s", actual, expected)
	}
}

func TestGenerateOSGInputTicketFile(t *testing.T) {
	cfg := InitConfig(t)
	s := InitTestsFromFile(t, cfg, "osg_submission.json")

	templateFields := OtherTemplateFields{TicketPathListHeader: cfg.GetString("tickets_path_list.file_identifier")}
	templateModel := TemplatesModel{s, templateFields}

	actual, err := generateInputTicketListContents(templateModel)
	if err != nil {
		t.Error(err)
	}
	expected := `# application/vnd.de.tickets-path-list+csv; version=1
1DD3D3EA-366B-41E6-A0CD-FFA915634E33,/iplant/home/sarahr/config.json
C4724423-159E-4391-B6B8-51AA686581A7,/iplant/home/sarahr/input_ticket.list
59F6624F-DF06-4E68-A6D5-125AB88F7919,/iplant/home/sarahr/output_ticket.list
`

	if actual.String() != expected {
		t.Errorf("generateInputTicketListContents() returned:\n\n%s\n\ninstead of\n\n%s", actual, expected)
	}
}

func TestGenerateOSGOutputTicketFile(t *testing.T) {
	cfg := InitConfig(t)
	s := InitTestsFromFile(t, cfg, "osg_submission.json")

	templateFields := OtherTemplateFields{TicketPathListHeader: cfg.GetString("tickets_path_list.file_identifier")}
	templateModel := TemplatesModel{s, templateFields}

	actual, err := generateOutputTicketListContents(templateModel)
	if err != nil {
		t.Error(err)
	}
	expected := `# application/vnd.de.tickets-path-list+csv; version=1
15AEE88E-A2B3-47E8-A862-2EBB77403A9F,/iplant/home/sarahr/analyses/osgwc_201804261317-2018-04-26-20-17-46.0
`

	if actual.String() != expected {
		t.Errorf("generateInputTicketListContents() returned:\n\n%s\n\ninstead of\n\n%s", actual, expected)
	}
}
