package jobs

import (
	"testing"
)

func TestGenerateOSGConfigJSONFile(t *testing.T) {
	cfg := InitConfig(t)
	s := InitTestsFromFile(t, cfg, "osg_submission.json")
	b := newOSGJobSubmissionBuilder(cfg)

	actual, err := generateJsonContents(b.generateConfig(s))
	if err != nil {
		t.Error(err)
	}

	expected := `{
    "irods_host": "wut.iplantc.org",
    "irods_port": 1247,
    "irods_job_user": "sarahr",
    "input_ticket_list": "input_ticket.list",
    "output_ticket_list": "output_ticket.list",
    "status_update_url": "http://sl.iplantc.org/2256dd6d-d984-4d3a-ad71-ab1ff341f636/status",
    "stdout": "out.txt",
    "stderr": "err.txt"
}
`

	if actual.String() != expected {
		t.Errorf("generateConfig() returned:\n\n%s\n\ninstead of:\n\n%s", actual, expected)
	}
}
