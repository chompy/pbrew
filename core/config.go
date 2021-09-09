package core

const brewDir = "homebrew"
const runDir = "run"
const confDir = "conf"
const dataDir = "data"
const varsDir = "vars"
const mntDir = "mnt"

// Config defines app configuration.
type Config struct {
	PortRangeStart int    `yaml:"port_range_start"`
	UserDir        string `yaml:"user_dir"`
	RouterHTTP     int    `yaml:"router_http_port"`
	RouterHTTPS    int    `yaml:"router_https_port"`
}

// DefaultConfig returns the default configuration settings.
func DefaultConfig() Config {
	return Config{
		PortRangeStart: 51000,
		UserDir:        "~/.pbrew",
		RouterHTTP:     80,
		RouterHTTPS:    443,
	}
}

// LoadConfig returns user configuration.
func LoadConfig() (Config, error) {
	config := DefaultConfig()
	// TODO
	return config, nil
}
