package core

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"github.com/pkg/errors"
)

// Service is a Homebrew service.
type Service struct {
	BrewName       string `yaml:"brew_name"`
	PostInstallCmd string `yaml:"post_install"`
	PostStartCmd   string `yaml:"post_start"`
	Port           int    `yaml:"port"`
}

// Info returns information about Homebrew application.
func (s *Service) Info() (map[string]interface{}, error) {
	binPath := filepath.Join(BrewPath(), "bin/brew")
	out, err := exec.Command(binPath, "info", s.BrewName, "--json").Output()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	info := make([]map[string]interface{}, 0)
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, errors.WithStack(err)
	}
	if len(info) == 0 {
		return nil, errors.WithStack(ErrServiceNotFound)
	}
	return info[0], nil
}

// Install installs the service and runs the post install command.
func (s *Service) Install() error {
	if err := brewCommand("install", s.BrewName); err != nil {
		return errors.WithStack(err)
	}
	if err := brewCommand("services", "stop", s.BrewName); err != nil {
		return errors.WithStack(err)
	}
	if err := s.PostInstall(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// PostInstall runs the post install command for the service.
func (s *Service) PostInstall() error {
	cmdStr := s.injectCommandParams(s.PostInstallCmd)
	if err := RunCommand(cmdStr); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// IsInstalled returns true if service is installed.
func (s *Service) IsInstalled() bool {
	info, err := s.Info()
	if err != nil {
		output.Error(err)
		return false
	}
	return len(info["installed"].([]interface{})) > 0
}

// Start will start the service.
func (s *Service) Start() error {
	if !s.IsInstalled() {
		return errors.WithStack(ErrServiceNotInstalled)
	}
	if err := brewCommand("services", "start", s.BrewName); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// PostStart runs the post start command for the service.
func (s *Service) PostStart(d interface{}) error {
	return nil
}

// Stop will stop the service.
func (s *Service) Stop() error {
	if !s.IsInstalled() {
		return errors.WithStack(ErrServiceNotInstalled)
	}
	if err := brewCommand("services", "stop", s.BrewName); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// SocketPath returns path to service socket.
func (s *Service) SocketPath() string {
	return fmt.Sprintf("/tmp/pbrew-%s.sock", s.BrewName)
}

func (s *Service) injectCommandParams(cmd string) string {
	cmd = strings.ReplaceAll(cmd, "{BREW_PATH}", BrewPath())
	cmd = strings.ReplaceAll(cmd, "{PORT}", fmt.Sprintf("%d", s.Port))
	cmd = strings.ReplaceAll(cmd, "{SOCKET}", s.SocketPath())
	return cmd
}
