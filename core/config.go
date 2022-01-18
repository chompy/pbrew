package core

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var loadedConfig *Config

// Config defines app configuration.
type Config struct {
	PortRangeStart   int               `yaml:"port_range_start"`
	UserDir          string            `yaml:"user_dir"`
	RouterHTTP       int               `yaml:"router_http_port"`
	RouterHTTPS      int               `yaml:"router_https_port"`
	Shell            string            `yaml:"shell"`
	ServiceOverrides []ServiceOverride `yaml:"service_overrides"`
}

// DefaultConfig returns the default configuration settings.
func DefaultConfig() Config {
	return Config{
		PortRangeStart:   51000,
		UserDir:          "~/.pbrew",
		RouterHTTP:       80,
		RouterHTTPS:      443,
		Shell:            "bash",
		ServiceOverrides: make([]ServiceOverride, 0),
	}
}

// LoadConfig returns user configuration.
func LoadConfig() (Config, error) {
	if loadedConfig != nil {
		return *loadedConfig, nil
	}
	config := DefaultConfig()
	configPath := filepath.Join(getAppPath(), "config.yaml")
	configRaw, err := ioutil.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		loadedConfig = &config
		return config, err
	}
	err = yaml.Unmarshal(configRaw, &config)
	loadedConfig = &config
	return config, err
}
