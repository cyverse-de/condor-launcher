package test

import (
	"fmt"
	"io"
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
	return path.Dir(sourcePath)
}

func getTestFilePath(t *testing.T, filename string) string {
	return path.Join(getTestConfigDir(t), filename)
}

// InitPath sets up the PATH environment variable for the tests.
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

// InitConfig sets up a config object for use in the tests.
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
	cfg.Set("condor.request_disk", "0")
	cfg.Set("condor.path_env_var", "/path/to/path")
	cfg.Set("condor.condor_config", "/condor/config")

	return cfg
}

// InitTestsFromFile loads test information from a file.
func InitTestsFromFile(t *testing.T, cfg *viper.Viper, filename string) *model.Job {

	// Load the job submission information from the file.
	path := getTestFilePath(t, filename)
	f, err := os.Open(path)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
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

// InitTests is basically an alias for InitTestsFromFile that defaults to
// loading info from test_submission.json.
func InitTests(t *testing.T, cfg *viper.Viper) *model.Job {
	return InitTestsFromFile(t, cfg, "test_submission.json")
}
