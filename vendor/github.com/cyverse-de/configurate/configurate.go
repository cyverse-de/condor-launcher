// Package configurate provides common configuration functionality.
// It supports reading configuration settings from a YAML file. All of the job
// services are intended to read from the same configuration file.
package configurate

import (
	"bytes"
	"io"
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

// InitDefaultsR initializes the underlying config from a reader, using defaults for any unspecified configuration
// settings.
func InitDefaultsR(reader io.Reader, defaultConfig string) (*viper.Viper, error) {
	cfg := viper.New()
	cfg.SetConfigType("yaml")

	// Load the default configuration.
	if err := cfg.ReadConfig(bytes.NewBuffer([]byte(defaultConfig))); err != nil {
		return nil, err
	}

	// Return the defaults if no configuration file was provided.
	if reader == nil {
		return cfg, nil
	}

	// Merge the configuration file settings.
	if err := cfg.MergeConfig(reader); err != nil {
		return nil, err
	}

	return cfg, nil
}

// InitDefaults initializes the underlying config, using defaults for any unspecified configuration settings.
func InitDefaults(path, defaultConfig string) (*viper.Viper, error) {

	// Open the configuration file. If the configuration file doesn't exist, simply return the default config.
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return InitDefaultsR(nil, defaultConfig)
		}
		return nil, err
	}
	defer f.Close()

	return InitDefaultsR(f, defaultConfig)
}
