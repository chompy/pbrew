package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"github.com/pkg/errors"
)

// Service is a Homebrew service.
type Service struct {
	BrewName       string `yaml:"brew_name"`
	PostInstallCmd string `yaml:"post_install"`
	StartCmd       string `yaml:"start"`
	StopCmd        string `yaml:"stop"`
	ReloadCmd      string `yaml:"reload"`
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

// IsRunning returns true if service is running.
func (s *Service) IsRunning() bool {
	pidFile, err := ioutil.ReadFile(s.PidPath())
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		output.Warn(err.Error())
		return false
	}
	pid, err := strconv.Atoi(string(bytes.TrimSpace(pidFile)))
	if err != nil {
		output.Warn(err.Error())
		return false
	}
	_, err = os.FindProcess(pid)
	return err == nil
}

// Start will start the service.
func (s *Service) Start() error {
	done := output.Duration(fmt.Sprintf("Start %s.", s.BrewName))
	if !s.IsInstalled() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotInstalled, s.BrewName))
	}
	if s.IsRunning() {
		return errors.WithStack(errors.WithMessage(ErrServiceAlreadyRunning, s.BrewName))
	}
	cmdStr := s.injectCommandParams(s.StartCmd)
	if err := RunCommand(cmdStr); err != nil {
		return errors.WithStack(err)
	}
	done()
	return nil
}

// Stop will stop the service.
func (s *Service) Stop() error {
	done := output.Duration(fmt.Sprintf("Stop %s.", s.BrewName))
	if !s.IsInstalled() {
		return errors.WithStack(ErrServiceNotInstalled)
	}
	if !s.IsRunning() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotRunning, s.BrewName))
	}

	cmdStr := s.injectCommandParams(s.StopCmd)
	if err := RunCommand(cmdStr); err != nil {
		return errors.WithStack(err)
	}
	done()
	return nil
}

// Reload reloads the service configuration.
func (s *Service) Reload() error {
	done := output.Duration(fmt.Sprintf("Reload %s.", s.BrewName))
	if s.ReloadCmd == "" {
		return errors.WithStack(errors.WithMessage(ErrServiceReloadNotDefined, s.BrewName))
	}
	if !s.IsInstalled() {
		return errors.WithStack(ErrServiceNotInstalled)
	}
	if !s.IsRunning() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotRunning, s.BrewName))
	}
	cmdStr := s.injectCommandParams(s.ReloadCmd)
	if err := RunCommand(cmdStr); err != nil {
		return errors.WithStack(err)
	}
	done()
	return nil
}

// SocketPath returns path to service socket.
func (s *Service) SocketPath() string {
	return fmt.Sprintf("/tmp/pbrew-%s.sock", s.BrewName)
}

// PidPath returns path to service pid file.
func (s *Service) PidPath() string {
	return fmt.Sprintf("/tmp/pbrew-%s.pid", strings.ReplaceAll(s.BrewName, "@", "-"))
}

func (s *Service) injectCommandParams(cmd string) string {
	cmd = strings.ReplaceAll(cmd, "{BREW_PATH}", BrewPath())
	cmd = strings.ReplaceAll(cmd, "{PORT}", fmt.Sprintf("%d", s.Port))
	cmd = strings.ReplaceAll(cmd, "{SOCKET}", s.SocketPath())
	cmd = strings.ReplaceAll(cmd, "{PID_FILE}", s.PidPath())
	cmd = strings.ReplaceAll(cmd, "{PID_FILE_ESC}", strings.ReplaceAll(s.PidPath(), "/", "\\/"))
	appPath, _ := appPath()
	cmd = strings.ReplaceAll(cmd, "{APP_PATH}", appPath)
	return cmd
}
