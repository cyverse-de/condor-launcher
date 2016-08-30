// Package configurate provides common configuration functionality.
// It supports reading configuration settings from a YAML file. All of the job
// services are intended to read from the same configuration file.
package configurate

import (
	"bytes"
	"os"

	"github.com/spf13/viper"
)

// Init initializes the underlying config.
func Init(path string) (*viper.Viper, error) {
	cfg := viper.New()
	cfg.SetConfigType("yaml")

	// Open the configuration file.
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Parse the configuration file contents.
	if err := cfg.ReadConfig(f); err != nil {
		return nil, err
	}

	return cfg, nil
}

// InitDefaults initializes the underlying config, using defaults for any unspecified configuration settings.
func InitDefaults(path, defaultConfig string) (*viper.Viper, error) {
	cfg := viper.New()
	cfg.SetConfigType("yaml")

	// Load the default configuration.
	if err := cfg.ReadConfig(bytes.NewBuffer([]byte(defaultConfig))); err != nil {
		return nil, err
	}

	// Open the configuration file. If the configuration file doesn't exist, simply return the default config.
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	defer f.Close()

	// Merge the configuration file settings.
	if err := cfg.MergeConfig(f); err != nil {
		return nil, err
	}

	return cfg, nil
}
