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
	"syscall"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"github.com/pkg/errors"
)

// Service is a Homebrew service.
type Service struct {
	BrewName       string `yaml:"brew_name"`
	PostInstallCmd string `yaml:"post_install"`
	ConfigTemplate string `yaml:"config_template"`
	StartCmd       string `yaml:"start"`
	StopCmd        string `yaml:"stop"`
	ReloadCmd      string `yaml:"reload"`
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
		return nil, errors.WithStack(errors.WithMessage(ErrServiceNotFound, s.BrewName))
	}
	return info[0], nil
}

// Install installs the service and runs the post install command.
func (s *Service) Install() error {
	if err := brewCommand("install", s.BrewName, "--force-bottle"); err != nil {
		return err
	}
	if err := brewCommand("services", "stop", s.BrewName); err != nil {
		return err
	}
	if err := s.PostInstall(); err != nil {
		return err
	}
	return nil
}

// PostInstall runs the post install command for the service.
func (s *Service) PostInstall() error {
	if s.PostInstallCmd != "" {
		cmdStr := s.injectCommandParams(s.PostInstallCmd)
		cmd := NewShellCommand()
		cmd.Args = []string{"-c", cmdStr}
		cmd.Env = os.Environ()
		if err := cmd.Interactive(); err != nil {
			return errors.WithMessage(err, s.BrewName)
		}
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
	proc, err := os.FindProcess(pid)
	if err != nil || proc == nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil || strings.Contains(err.Error(), "not permitted")
}

// Start will start the service.
func (s *Service) Start() error {
	done := output.Duration(fmt.Sprintf("Start %s.", s.BrewName))
	// check status
	if !s.IsInstalled() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotInstalled, s.BrewName))
	}
	if s.IsRunning() {
		return errors.WithStack(errors.WithMessage(ErrServiceAlreadyRunning, s.BrewName))
	}
	// generate config file
	if err := s.GenerateConfigFile(); err != nil {
		return err
	}
	// create data dir
	if err := os.Mkdir(s.DataPath(), mkdirPerm); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return errors.WithStack(err)
		}
	}
	// execute start cmd
	done2 := output.Duration("Start up.")
	cmdStr := s.injectCommandParams(s.StartCmd)
	cmd := NewShellCommand()
	cmd.Args = []string{"-c", cmdStr}
	cmd.Env = os.Environ()
	if err := cmd.Interactive(); err != nil {
		return errors.WithMessage(err, s.BrewName)
	}
	done2()
	done()
	return nil
}

// Stop will stop the service.
func (s *Service) Stop() error {
	done := output.Duration(fmt.Sprintf("Stop %s.", s.BrewName))
	// check status
	if !s.IsInstalled() {
		return errors.WithStack(ErrServiceNotInstalled)
	}
	if !s.IsRunning() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotRunning, s.BrewName))
	}
	// execute stop cmd
	cmdStr := s.injectCommandParams(s.StopCmd)
	cmd := NewShellCommand()
	cmd.Args = []string{"-c", cmdStr}
	cmd.Env = os.Environ()
	if err := cmd.Interactive(); err != nil {
		return errors.WithMessage(err, s.BrewName)
	}
	done()
	return nil
}

// Reload reloads the service configuration.
func (s *Service) Reload() error {
	done := output.Duration(fmt.Sprintf("Reload %s.", s.BrewName))
	// check status
	if s.ReloadCmd == "" {
		return errors.WithStack(errors.WithMessage(ErrServiceReloadNotDefined, s.BrewName))
	}
	if !s.IsInstalled() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotInstalled, s.BrewName))
	}
	if !s.IsRunning() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotRunning, s.BrewName))
	}
	// execute reload cmd
	cmdStr := s.injectCommandParams(s.ReloadCmd)
	cmd := NewShellCommand()
	cmd.Args = []string{"-c", cmdStr}
	cmd.Env = os.Environ()
	if err := cmd.Interactive(); err != nil {
		return errors.WithMessage(err, s.BrewName)
	}
	done()
	return nil
}

// PreStart performs setup that should occur prior to starting service.
func (s *Service) PreStart(d interface{}, p *Project) error {
	switch d := d.(type) {
	case *def.App:
		{
			done := output.Duration(fmt.Sprintf("Pre setup %s.", d.Name))
			if s.IsRunning() {
				return errors.WithStack(errors.WithMessage(ErrServiceAlreadyRunning, s.BrewName))
			}
			if s.IsPHP() {
				if err := s.phpPreSetup(d, p); err != nil {
					return err
				}
			}
			done()
			break
		}
	}
	return nil
}

// PostStart performs setup that should occur after starting service.
func (s *Service) PostStart(d interface{}, p *Project) error {
	switch d := d.(type) {
	case *def.Service:
		{
			done := output.Duration(fmt.Sprintf("Post setup %s.", d.Name))
			if !s.IsRunning() {
				return errors.WithStack(errors.WithMessage(ErrServiceNotRunning, s.BrewName))
			}
			if s.IsMySQL() {
				if err := s.mySQLPostSetup(d, p); err != nil {
					return err
				}
			}
			done()
			break
		}
	}
	return nil
}

// Cleanup performs cleanup for project.
func (s *Service) Cleanup(d interface{}, p *Project) error {
	switch d := d.(type) {
	case *def.App:
		{
			done := output.Duration(fmt.Sprintf("Clean up %s.", d.Name))
			if s.IsPHP() {
				if err := s.phpCleanup(d, p); err != nil {
					return err
				}
			}
			done()
			break
		}
	}
	return nil
}

// Purge deletes all data related to given project and service definition.
func (s *Service) Purge(d interface{}, p *Project) error {
	return nil
}

// Port returns the assigned port.
func (s *Service) Port() (int, error) {
	portMap, err := LoadPortMap()
	if err != nil {
		return 0, err
	}
	return portMap.ServicePort(s)
}

// SocketPath returns path to service socket.
func (s *Service) SocketPath() string {
	return filepath.Join(userPath(), runDir, fmt.Sprintf("%s.sock", strings.ReplaceAll(s.BrewName, "@", "-")))
}

// UpstreamSocketPath returns path to app upstream socket.
func (s *Service) UpstreamSocketPath(p *Project, app *def.App) string {
	if s.IsPHP() {
		return filepath.Join(userPath(), runDir, fmt.Sprintf("php-%s-%s.sock", p.Name, app.Name))
	}
	return s.SocketPath()
}

// PidPath returns path to service pid file.
func (s *Service) PidPath() string {
	return filepath.Join(userPath(), runDir, fmt.Sprintf("%s.pid", strings.ReplaceAll(s.BrewName, "@", "-")))
}

// ConfigPath returns path to service config file.
func (s *Service) ConfigPath() string {
	return filepath.Join(userPath(), confDir, fmt.Sprintf("%s.conf", strings.ReplaceAll(s.BrewName, "@", "-")))
}

// DataPath returns path to service data directory.
func (s *Service) DataPath() string {
	return filepath.Join(userPath(), dataDir, strings.ReplaceAll(s.BrewName, "@", "-"))
}

func (s *Service) injectCommandParams(cmd string) string {
	port, err := s.Port()
	if err != nil {
		output.Warn(err.Error())
	}
	cmd = strings.ReplaceAll(cmd, "{BREW_PATH}", BrewPath())
	cmd = strings.ReplaceAll(cmd, "{BREW_APP}", s.BrewName)
	cmd = strings.ReplaceAll(cmd, "{PORT}", fmt.Sprintf("%d", port))
	cmd = strings.ReplaceAll(cmd, "{SOCKET}", s.SocketPath())
	cmd = strings.ReplaceAll(cmd, "{PID_FILE}", s.PidPath())
	cmd = strings.ReplaceAll(cmd, "{PID_FILE_ESC}", strings.ReplaceAll(s.PidPath(), "/", "\\/"))
	appPath, _ := appPath()
	cmd = strings.ReplaceAll(cmd, "{APP_PATH}", appPath)
	cmd = strings.ReplaceAll(cmd, "{CONF_FILE}", s.ConfigPath())
	cmd = strings.ReplaceAll(cmd, "{CONF_PATH}", filepath.Join(userPath(), confDir))
	cmd = strings.ReplaceAll(cmd, "{DATA_PATH}", s.DataPath())
	return cmd
}
