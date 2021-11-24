package core

import (
	"path/filepath"
	"strings"
)

// ServicesEnv returns environment variables for given services.
func ServicesEnv(services []*Service) []string {
	envPath := make([]string, 0)
	for _, service := range services {
		envPath = append(envPath, filepath.Join(GetDir(BrewDir), "opt", service.BrewAppName(), "bin"))
		for _, dependency := range service.Dependencies {
			envPath = append(envPath, filepath.Join(GetDir(BrewDir), "opt", dependency, "bin"))
		}
	}
	envPath = append(envPath, filepath.Join(GetDir(BrewDir), "bin"))
	envPath = append(envPath, "/bin")
	envPath = append(envPath, "/usr/bin")
	envPath = append(envPath, "/usr/sbin")
	env := make([]string, 0)
	env = append(env, brewEnv()...)
	for k, v := range env {
		if strings.HasPrefix(v, "PATH") {
			env = append(env[:k], env[k+1:]...)
			break
		}
	}
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
