package jobs

import (
	"testing"
)

func runOSGConfigTest(t *testing.T, submissionPath, expected string) {
	cfg := InitConfig(t)
	s := InitTestsFromFile(t, cfg, submissionPath)
	b := newOSGJobSubmissionBuilder(cfg)

	actual, err := generateJSONContents(b.generateConfig(s))
	if err != nil {
		t.Error(err)
	}

	if actual.String() != expected {
		t.Errorf("generateConfig() returned:\n\n%s\n\ninstead of:\n\n%s", actual, expected)
	}
}

func TestGenerateOSGConfigJSONFile(t *testing.T) {
	expected := `{
    "arguments": [],
    "irods_host": "wut.iplantc.org",
    "irods_port": 1247,
    "irods_job_user": "sarahr",
    "irods_user_name": "iuser",
    "irods_zone_name": "",
    "input_ticket_list": "input_ticket.list",
    "output_ticket_list": "output_ticket.list",
    "status_update_url": "http://sl.iplantc.org/2256dd6d-d984-4d3a-ad71-ab1ff341f636/status",
    "stdout": "out.txt",
    "stderr": "err.txt"
}
`

	runOSGConfigTest(t, "osg_submission.json", expected)
}

func TestGeneratedOSGConfigJSONFileArgs(t *testing.T) {
	expected := `{
    "arguments": [
        "-b"
    ],
    "irods_host": "wut.iplantc.org",
    "irods_port": 1247,
    "irods_job_user": "sarahr",
    "irods_user_name": "iuser",
    "irods_zone_name": "",
    "input_ticket_list": "input_ticket.list",
    "output_ticket_list": "output_ticket.list",
    "status_update_url": "http://sl.iplantc.org/2256dd6d-d984-4d3a-ad71-ab1ff341f636/status",
    "stdout": "out.txt",
    "stderr": "err.txt"
}
`

	runOSGConfigTest(t, "osg_submission_args.json", expected)
}
