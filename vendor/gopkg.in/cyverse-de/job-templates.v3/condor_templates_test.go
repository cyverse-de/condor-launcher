package jobs

import (
	"testing"
	"github.com/cyverse-de/condor-launcher/test"
)

func TestGenerateCondorSubmit(t *testing.T) {
	cfg := test.InitConfig(t)
	s := test.InitTests(t, cfg)

	actual, err := generateFileContents(condorSubmissionTemplate, s)
	if err != nil {
		t.Error(err)
	}
	expected := `universe = vanilla
executable = /usr/local/bin/road-runner
rank = mips
requirements = (HAS_HOST_MOUNTS == True)
arguments = --config config --job job
output = script-output.log
error = script-error.log
log = condor.log
accounting_group = de
accounting_group_user = test_this_is_a_test
request_disk = 0
+IpcUuid = "07b04ce2-7757-4b21-9e15-0b4c2f44be26"
+IpcJobId = "generated_script"
+IpcUsername = "test_this_is_a_test"
+IpcUserGroups = {"groups:foo","groups:bar","groups:baz"}
concurrency_limits = _00000000000000000000000000000000
+IpcExe = "wc_wrapper.sh"
+IpcExePath = "/usr/local3/bin/wc_tool-1.00"
should_transfer_files = YES
transfer_input_files = iplant.cmd,config,job
transfer_output_files = workingvolume/logs/logs-stdout-output,workingvolume/logs/logs-stderr-output
when_to_transfer_output = ON_EXIT_OR_EVICT
notification = NEVER
queue
`
	if actual.String() != expected {
		t.Errorf("GenerateCondorSubmit() returned:\n\n%s\n\ninstead of:\n\n%s", actual, expected)
	}
}

func TestGenerateCondorSubmitGroup(t *testing.T) {
	cfg := test.InitConfig(t)
	s := test.InitTests(t, cfg)
	s.Group = "foo"
	actual, err := generateFileContents(condorSubmissionTemplate, s)
	if err != nil {
		t.Error(err)
	}
	expected := `universe = vanilla
executable = /usr/local/bin/road-runner
rank = mips
requirements = (HAS_HOST_MOUNTS == True)
arguments = --config config --job job
output = script-output.log
error = script-error.log
log = condor.log
accounting_group = foo
accounting_group_user = test_this_is_a_test
request_disk = 0
+IpcUuid = "07b04ce2-7757-4b21-9e15-0b4c2f44be26"
+IpcJobId = "generated_script"
+IpcUsername = "test_this_is_a_test"
+IpcUserGroups = {"groups:foo","groups:bar","groups:baz"}
concurrency_limits = _00000000000000000000000000000000
+IpcExe = "wc_wrapper.sh"
+IpcExePath = "/usr/local3/bin/wc_tool-1.00"
should_transfer_files = YES
transfer_input_files = iplant.cmd,config,job
transfer_output_files = workingvolume/logs/logs-stdout-output,workingvolume/logs/logs-stderr-output
when_to_transfer_output = ON_EXIT_OR_EVICT
notification = NEVER
queue
`
	if actual.String() != expected {
		t.Errorf("GenerateCondorSubmit() returned:\n\n%s\n\ninstead of:\n\n%s", actual, expected)
	}
}

func TestGenerateCondorSubmitNoVolumes(t *testing.T) {
	cfg := test.InitConfig(t)
	s := test.InitTestsFromFile(t, cfg, "no_volumes_submission.json")
	actual, err := generateFileContents(condorSubmissionTemplate, s)
	if err != nil {
		t.Error(err)
	}
	expected := `universe = vanilla
executable = /usr/local/bin/road-runner
rank = mips
arguments = --config config --job job
output = script-output.log
error = script-error.log
log = condor.log
accounting_group = de
accounting_group_user = test_this_is_a_test
request_disk = 0
+IpcUuid = "07b04ce2-7757-4b21-9e15-0b4c2f44be26"
+IpcJobId = "generated_script"
+IpcUsername = "test_this_is_a_test"
+IpcUserGroups = {"groups:foo","groups:bar","groups:baz"}
concurrency_limits = _00000000000000000000000000000000
+IpcExe = "wc_wrapper.sh"
+IpcExePath = "/usr/local3/bin/wc_tool-1.00"
should_transfer_files = YES
transfer_input_files = iplant.cmd,config,job
transfer_output_files = workingvolume/logs/logs-stdout-output,workingvolume/logs/logs-stderr-output
when_to_transfer_output = ON_EXIT_OR_EVICT
notification = NEVER
queue
`
	if actual.String() != expected {
		t.Errorf("GenerateCondorSubmit() returned:\n\n%s\n\ninstead of:\n\n%s", actual, expected)
	}
}
