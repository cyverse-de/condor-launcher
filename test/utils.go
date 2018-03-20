package test

import (
	"testing"
	"github.com/cyverse-de/configurate"
	"fmt"
	"os"
	"gopkg.in/cyverse-de/model.v1"
	"github.com/spf13/viper"
	"io/ioutil"
	"runtime"
	"path"
	"strings"
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
	cfg.Set("condor.request_disk", "0")
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
