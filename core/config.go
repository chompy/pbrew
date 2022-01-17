package core

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config defines app configuration.
type Config struct {
	PortRangeStart int    `yaml:"port_range_start"`
	UserDir        string `yaml:"user_dir"`
	RouterHTTP     int    `yaml:"router_http_port"`
	RouterHTTPS    int    `yaml:"router_https_port"`
	Shell          string `yaml:"shell"`
}

// DefaultConfig returns the default configuration settings.
func DefaultConfig() Config {
	return Config{
		PortRangeStart: 51000,
		UserDir:        "~/.pbrew",
		RouterHTTP:     80,
		RouterHTTPS:    443,
		Shell:          "bash",
	}
}

// LoadConfig returns user configuration.
func LoadConfig() (Config, error) {
	config := DefaultConfig()
	configPath := filepath.Join(getAppPath(), "config.yaml")
	configRaw, err := ioutil.ReadFile(configPath)
	if err != nil {
		if err == os.ErrNotExist {
			return config, nil
		}
		return config, err
	}
	err = yaml.Unmarshal(configRaw, &config)
	return config, err
}
