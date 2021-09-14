package core

import (
	"path/filepath"
	"strings"
)

// ServicesEnv returns environment variables for given services.
func ServicesEnv(services []*Service) []string {
	envPath := make([]string, 0)
	envPath = append(envPath, filepath.Join(GetDir(BrewDir), "bin"))
	for _, service := range services {
		envPath = append(envPath, filepath.Join(GetDir(BrewDir), "opt", service.BrewName, "bin"))
	}
	envPath = append(envPath, "/bin")
	envPath = append(envPath, "/usr/bin")
	env := make([]string, 0)
	env = append(env, "PATH="+strings.Join(envPath, ":"))
	env = append(env, "HOME="+GetDir(HomeDir))
	for _, service := range services {
		if service.IsPHP() {
			env = append(env, "PHPRC="+service.DataPath())
			break
		}
	}
	return env
}
