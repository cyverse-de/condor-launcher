package configurate

import (
	"testing"
)

const cfgFile = "test/test_config.yaml"
const partialCfgFile = "test/test_partial_config.yaml"
const missingCfgFile = "test/ekky_ekky_ekky_zboing_zoom_bang_znourringmn.yaml"

func TestNew(t *testing.T) {
	cfg, err := Init(cfgFile)
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Errorf("configurate.New() returned nil")
	}
}

func TestConfig(t *testing.T) {
	cfg, err := Init(cfgFile)
	if err != nil {
		t.Fatal(err)
	}

	// Spot-check a configuration setting.
	actual := cfg.GetString("amqp.uri")
	expected := "amqp://guest:guest@rabbit:5672/jobs"
	if actual != expected {
		t.Errorf("The amqp.uri was %s instead of %s", actual, expected)
	}

	// Spot-check another configuration setting.
	actual = cfg.GetString("condor.amqp.launches_queue.name")
	expected = "condor_launchers"
	if actual != expected {
		t.Errorf("The condor.amqp.launches_queue.name was %s instead of %s", actual, expected)
	}
}

func TestDefaults(t *testing.T) {
	cfg, err := InitDefaults(partialCfgFile, JobServicesDefaults)
	if err != nil {
		t.Fatal(err)
	}

	// Spot-check an unspecified configuration setting.
	actual := cfg.GetString("apps.callbacks_uri")
	expected := "http://apps:60000/callbacks/de-job"
	if actual != expected {
		t.Errorf("The defaulted apps.callbacks_uri was %s instead of %s", actual, expected)
	}

	// Spot-check an overridden configuration setting.
	actual = cfg.GetString("amqp.uri")
	expected = "amqp://guest:guest@wabbit:5672/fudd"
	if actual != expected {
		t.Errorf("The overridden amqp.uri was %s instead of %s", actual, expected)
	}

	// Spot-check an overridden value in an object.
	actual = cfg.GetString("irods.pass")
	expected = "prodnaught"
	if actual != expected {
		t.Errorf("The overriden irods.pass was %s instead of %s", actual, expected)
	}

	// Spot-check a non-overridden value in an object.
	actual = cfg.GetString("irods.host")
	expected = "irods"
	if actual != expected {
		t.Errorf("The defaulted irods.host was %s instead of %s", actual, expected)
	}
}

func TestDefaultsWithNonExistentConfigFile(t *testing.T) {
	cfg, err := InitDefaults(missingCfgFile, JobServicesDefaults)
	if err != nil {
		t.Fatal(err)
	}

	// Spot-check a configuration setting.
	actual := cfg.GetString("apps.callbacks_uri")
	expected := "http://apps:60000/callbacks/de-job"
	if actual != expected {
		t.Errorf("The defaulted apps.callbacks_uri was %s instead of %s", actual, expected)
	}

	// Spot-check another configuration setting for good measure.
	actual = cfg.GetString("amqp.uri")
	expected = "amqp://guest:guest@rabbit:5672/jobs"
	if actual != expected {
		t.Errorf("The overridden amqp.uri was %s instead of %s", actual, expected)
	}
}
