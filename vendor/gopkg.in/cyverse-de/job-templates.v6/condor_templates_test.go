package jobs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/cyverse-de/configurate"
	"github.com/spf13/viper"
	"gopkg.in/cyverse-de/model.v4"
)

func getTestConfigDir(t *testing.T) string {
	_, sourcePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Errorf("unable to determine the path to the current source file")
	}
	return path.Join(path.Dir(sourcePath), "test")
}

func getTestFilePath(t *testing.T, filename string) string {
	return path.Join(getTestConfigDir(t), filename)
}

func InitPath(t *testing.T) {
	dir := getTestConfigDir(t)
	path := os.Getenv("PATH")

	// Don't bother updating the path if the test config directory is already in it.
	if strings.HasPrefix(path, fmt.Sprintf("%s:", dir)) {
		return
	}

	// Update the path.
	err := os.Setenv("PATH", fmt.Sprintf("%s:%s", dir, path))
	if err != nil {
		t.Error(err)
	}
}

func InitConfig(t *testing.T) *viper.Viper {

	// Load the test configuration.
	path := getTestFilePath(t, "test_config.yaml")
	cfg, err := configurate.InitDefaults(path, configurate.JobServicesDefaults)
	if err != nil {
		t.Error(err)
	}

	// Set test configuration values.
	cfg.Set("irods.base", "/path/to/irodsbase")
	cfg.Set("irods.host", "hostname")
	cfg.Set("irods.port", "1247")
	cfg.Set("irods.user", "user")
	cfg.Set("irods.pass", "pass")
	cfg.Set("irods.zone", "test")
	cfg.Set("irods.resc", "")
	cfg.Set("condor.log_path", "test")
	cfg.Set("condor.porklock_tag", "test")
	cfg.Set("condor.filter_files", "foo,bar,baz,blippy")
	cfg.Set("condor.path_env_var", "/path/to/path")
	cfg.Set("condor.condor_config", "/condor/config")
	cfg.Set("vault.irods.child_token.token", "token")
	cfg.Set("vault.irods.child_token.use_limit", 3)
	cfg.Set("vault.irods.mount_path", "irods")

	return cfg
}

func InitTestsFromFile(t *testing.T, cfg *viper.Viper, filename string) *model.Job {

	// Load the job submission information from the file.
	path := getTestFilePath(t, filename)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error(err)
	}

	// Parse the job submission information.
	s, err := model.NewFromData(cfg, data)
	if err != nil {
		t.Error(err)
	}

	// Add the test configuration directory to the PATH.
	InitPath(t)

	return s
}

func InitTests(t *testing.T, cfg *viper.Viper) *model.Job {
	return InitTestsFromFile(t, cfg, "test_submission.json")
}

func TestGenerateCondorSubmit(t *testing.T) {
	cfg := InitConfig(t)
	s := InitTests(t, cfg)

	actual, err := generateFileContents(condorSubmissionTemplate, s)
	if err != nil {
		t.Error(err)
	}
	expected := `universe = vanilla
executable = /usr/local/bin/road-runner
rank = 100 - TotalLoadAvg
requirements = HasDocker && (HAS_CYVERSE_ROAD_RUNNER =?= True) && (HAS_HOST_MOUNTS =?= True)
request_cpus = 4.5
request_memory = 2048MB
request_disk = 2048MB
arguments = --config config --job job
output = script-output.log
error = script-error.log
log = condor.log
accounting_group = de
accounting_group_user = test_this_is_a_test
+IpcUuid = "07b04ce2-7757-4b21-9e15-0b4c2f44be26"
+IpcJobId = "generated_script"
+IpcUsername = "test_this_is_a_test"
+IpcUserGroups = {"groups:foo","groups:bar","groups:baz"}
concurrency_limits = _00000000000000000000000000000000
+IpcExe = "wc_wrapper.sh"
+IpcExePath = "/usr/local3/bin/wc_tool-1.00"
should_transfer_files = YES
transfer_input_files = irods-config,iplant.cmd,config,job
transfer_output_files = workingvolume/logs/logs-stdout-output,workingvolume/logs/logs-stderr-output
when_to_transfer_output = ON_EXIT_OR_EVICT
notification = NEVER
queue
`
	if actual.String() != expected {
		t.Errorf("GenerateCondorSubmit() returned:\n\n%s\n\ninstead of:\n\n%s", actual, expected)
	}
}

func TestGenerateCondorSubmitExtraRequirements(t *testing.T) {
	cfg := InitConfig(t)
	s := InitTestsFromFile(t, cfg, "extra_requirements_submission.json")
	actual, err := generateFileContents(condorSubmissionTemplate, s)
	if err != nil {
		t.Error(err)
	}

	expected := `universe = vanilla
executable = /usr/local/bin/road-runner
rank = 100 - TotalLoadAvg
requirements = HasDocker && (HAS_CYVERSE_ROAD_RUNNER =?= True) && (HAS_HOST_MOUNTS =?= True) && (TRUE)
request_cpus = 4.5
request_memory = 2048MB
request_disk = 2048MB
arguments = --config config --job job
output = script-output.log
error = script-error.log
log = condor.log
accounting_group = de
accounting_group_user = test_this_is_a_test
+IpcUuid = "07b04ce2-7757-4b21-9e15-0b4c2f44be26"
+IpcJobId = "generated_script"
+IpcUsername = "test_this_is_a_test"
+IpcUserGroups = {"groups:foo","groups:bar","groups:baz"}
concurrency_limits = _00000000000000000000000000000000
+IpcExe = "wc_wrapper.sh"
+IpcExePath = "/usr/local3/bin/wc_tool-1.00"
should_transfer_files = YES
transfer_input_files = irods-config,iplant.cmd,config,job
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
	cfg := InitConfig(t)
	s := InitTests(t, cfg)
	s.Group = "foo"
	actual, err := generateFileContents(condorSubmissionTemplate, s)
	if err != nil {
		t.Error(err)
	}
	expected := `universe = vanilla
executable = /usr/local/bin/road-runner
rank = 100 - TotalLoadAvg
requirements = HasDocker && (HAS_CYVERSE_ROAD_RUNNER =?= True) && (HAS_HOST_MOUNTS =?= True)
request_cpus = 4.5
request_memory = 2048MB
request_disk = 2048MB
arguments = --config config --job job
output = script-output.log
error = script-error.log
log = condor.log
accounting_group = foo
accounting_group_user = test_this_is_a_test
+IpcUuid = "07b04ce2-7757-4b21-9e15-0b4c2f44be26"
+IpcJobId = "generated_script"
+IpcUsername = "test_this_is_a_test"
+IpcUserGroups = {"groups:foo","groups:bar","groups:baz"}
concurrency_limits = _00000000000000000000000000000000
+IpcExe = "wc_wrapper.sh"
+IpcExePath = "/usr/local3/bin/wc_tool-1.00"
should_transfer_files = YES
transfer_input_files = irods-config,iplant.cmd,config,job
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
	cfg := InitConfig(t)
	s := InitTestsFromFile(t, cfg, "no_volumes_submission.json")
	actual, err := generateFileContents(condorSubmissionTemplate, s)
	if err != nil {
		t.Error(err)
	}
	expected := `universe = vanilla
executable = /usr/local/bin/road-runner
rank = 100 - TotalLoadAvg
requirements = HasDocker && (HAS_CYVERSE_ROAD_RUNNER =?= True)
request_memory = 2KB
arguments = --config config --job job
output = script-output.log
error = script-error.log
log = condor.log
accounting_group = de
accounting_group_user = test_this_is_a_test
+IpcUuid = "07b04ce2-7757-4b21-9e15-0b4c2f44be26"
+IpcJobId = "generated_script"
+IpcUsername = "test_this_is_a_test"
+IpcUserGroups = {"groups:foo","groups:bar","groups:baz"}
concurrency_limits = _00000000000000000000000000000000
+IpcExe = "wc_wrapper.sh"
+IpcExePath = "/usr/local3/bin/wc_tool-1.00"
should_transfer_files = YES
transfer_input_files = irods-config,iplant.cmd,config,job
transfer_output_files = workingvolume/logs/logs-stdout-output,workingvolume/logs/logs-stderr-output
when_to_transfer_output = ON_EXIT_OR_EVICT
notification = NEVER
queue
`
	if actual.String() != expected {
		t.Errorf("GenerateCondorSubmit() returned:\n\n%s\n\ninstead of:\n\n%s", actual, expected)
	}
}

func TestCondorBytes(t *testing.T) {
	cases := []struct {
		input  int64
		output string
	}{
		{input: 400, output: "1KB"},
		{input: 1024, output: "1KB"},
		{input: 1000000, output: "977KB"},
		{input: 1048576, output: "1MB"},
		{input: 1000000000, output: "954MB"},
		{input: 16000000000, output: "15259MB"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%d", c.input), func(t *testing.T) {
			output := CondorBytes(c.input)
			if output != c.output {
				t.Errorf("Got %s from input %d, expected %s", output, c.input, c.output)
			}
		})
	}
}
