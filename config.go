package main

import "github.com/spf13/viper"

// CopyConfig will create a new *viper.Viper containing all of the settings
// from the *viper.Viper passed in. Updating one will not update the other.
func CopyConfig(cfg *viper.Viper) *viper.Viper {
	new := viper.New()
	for k, v := range cfg.AllSettings() {
		new.Set(k, v)
	}
	return new
}
