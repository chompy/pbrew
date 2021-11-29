package core

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
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
	BrewName        string            `yaml:"brew_name"`
	PostInstallCmd  string            `yaml:"post_install"`
	PreInstallCmd   string            `yaml:"pre_install"`
	ConfigTemplates map[string]string `yaml:"config_templates"`
	StartCmd        string            `yaml:"start"`
	StopCmd         string            `yaml:"stop"`
	ReloadCmd       string            `yaml:"reload"`
	InstallCheckCmd string            `yaml:"install_check"`
	Dependencies    []string          `yaml:"dependencies"`
}

// Info returns information about Homebrew application.
func (s *Service) Info() (map[string]interface{}, error) {
	return brewInfo(s.BrewAppName())
}

// Install installs the service and runs the post install command.
func (s *Service) Install() error {
	if err := s.PreInstall(); err != nil {
		return err
	}
	installName := s.BrewName
	if err := brewBottleDownload(s.BrewName); err == nil {
		installName = brewBottleDownloadPath(s.BrewName)
	}
	if err := brewCommand("install", installName); err != nil {
		if !s.InstallCheck() {
			return err
		}
		output.Warn("Install command errored but install check was successful... " + err.Error())
	}
	if err := brewCommand("services", "stop", s.BrewName); err != nil {
		return err
	}
	if err := s.PostInstall(); err != nil {
		return err
	}
	return nil
}

// InstallCheck checks to see if the installation was successful.
func (s *Service) InstallCheck() bool {
	// run cmd
	if s.InstallCheckCmd != "" {
		cmdStr := s.injectCommandParams(s.InstallCheckCmd)
		cmd := NewShellCommand()
		cmd.Args = []string{"--norc", "-c", cmdStr}
		cmd.Env = ServicesEnv([]*Service{s})
		if err := cmd.Interactive(); err != nil {
			return false
		}
	}
	return true
}

// PreInstall runs the pre install command for the service.
func (s *Service) PreInstall() error {
	// run cmd
	if s.PreInstallCmd != "" {
		cmdStr := s.injectCommandParams(s.PreInstallCmd)
		cmd := NewShellCommand()
		cmd.Args = []string{"--norc", "-c", cmdStr}
		cmd.Env = ServicesEnv([]*Service{s})
		if err := cmd.Interactive(); err != nil {
			return errors.WithMessage(err, s.BrewName)
		}
	}
	return nil
}

// PostInstall runs the post install command for the service.
func (s *Service) PostInstall() error {
	// create data dir
	if err := os.Mkdir(s.DataPath(), mkdirPerm); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return errors.WithStack(err)
		}
	}
	// run cmd
	if s.PostInstallCmd != "" {
		cmdStr := s.injectCommandParams(s.PostInstallCmd)
		cmd := NewShellCommand()
		cmd.Args = []string{"--norc", "-c", cmdStr}
		cmd.Env = ServicesEnv([]*Service{s})
		if err := cmd.Interactive(); err != nil {
			return errors.WithMessage(err, s.BrewName)
		}
	}
	return nil
}

// InstallDependencies installs brew dependencies for given service.
func (s *Service) InstallDependencies() error {
	for _, name := range s.Dependencies {
		dependService := Service{
			BrewName: name,
		}
		if dependService.IsInstalled() {
			continue
		}
		done := output.Duration(fmt.Sprintf("Install dependency %s.", name))
		if err := dependService.Install(); err != nil {
			return err
		}
		done()
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
	// execute start cmd
	done2 := output.Duration("Start up.")
	cmdStr := s.injectCommandParams(s.StartCmd)
	cmd := NewShellCommand()
	cmd.Args = []string{"--norc", "-c", cmdStr}
	cmd.Env = ServicesEnv([]*Service{s})
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
	cmd.Args = []string{"--norc", "-c", cmdStr}
	cmd.Env = ServicesEnv([]*Service{s})
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
	cmd.Args = []string{"--norc", "-c", cmdStr}
	cmd.Env = ServicesEnv([]*Service{s})
	if err := cmd.Interactive(); err != nil {
		return errors.WithMessage(err, s.BrewName)
	}
	done()
	return nil
}

// PreStart performs setup that should occur prior to starting service.
func (s *Service) PreStart(d interface{}, p *Project) error {
	done := output.Duration(fmt.Sprintf("Pre setup %s.", s.BrewName))
	// create data dir
	if err := os.Mkdir(s.DataPath(), mkdirPerm); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return errors.WithStack(err)
		}
	}
	// generate config file
	if err := s.GenerateConfigFile(); err != nil {
		return err
	}
	// service specific setup
	switch d := d.(type) {
	case *def.App:
		{
			if s.IsPHP() {
				if err := s.phpPreSetup(d, p); err != nil {
					return err
				}
			}
			break
		}
	}
	done()
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
	switch d := d.(type) {
	case *def.Service:
		{
			done := output.Duration(fmt.Sprintf("Purge %s.", d.Name))
			if s.IsMySQL() {
				if err := s.mySQLPurge(d, p); err != nil {
					return err
				}
			}
			done()
			break
		}
	}
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
	return filepath.Join(GetDir(RunDir), fmt.Sprintf("%s.sock", strings.ReplaceAll(s.BrewAppName(), "@", "-")))
}

// UpstreamSocketPath returns path to app upstream socket.
func (s *Service) UpstreamSocketPath(p *Project, app *def.App) string {
	if s.IsPHP() {
		return filepath.Join(GetDir(RunDir), fmt.Sprintf("php-%s-%s.sock", p.Name, app.Name))
	}
	return s.SocketPath()
}

// BrewAppName returns the brew app name without namespace.
func (s *Service) BrewAppName() string {
	return brewAppName(s.BrewName)
}

// PidPath returns path to service pid file.
func (s *Service) PidPath() string {
	return filepath.Join(GetDir(RunDir), fmt.Sprintf("%s.pid", strings.ReplaceAll(s.BrewAppName(), "@", "-")))
}

// ConfigPath returns path to service config file.
func (s *Service) ConfigPath() string {
	return filepath.Join(GetDir(ConfDir), fmt.Sprintf("%s.conf", strings.ReplaceAll(s.BrewAppName(), "@", "-")))
}

// DataPath returns path to service data directory.
func (s *Service) DataPath() string {
	return filepath.Join(GetDir(DataDir), strings.ReplaceAll(s.BrewAppName(), "@", "-"))
}

// ConfigParams returns confir template parameters for service.
func (s *Service) ConfigParams() map[string]interface{} {
	if s.IsPHP() {
		return s.phpConfigParams()
	}
	return map[string]interface{}{}
}

func (s *Service) injectCommandParams(cmd string) string {
	port, err := s.Port()
	if err != nil {
		output.Warn(err.Error())
	}

	cmd = strings.ReplaceAll(cmd, "{BREW_PATH}", GetDir(BrewDir))
	cmd = strings.ReplaceAll(cmd, "{BREW_APP}", s.BrewAppName())
	cmd = strings.ReplaceAll(cmd, "{PORT}", fmt.Sprintf("%d", port))
	cmd = strings.ReplaceAll(cmd, "{SOCKET}", s.SocketPath())
	cmd = strings.ReplaceAll(cmd, "{PID_FILE}", s.PidPath())
	cmd = strings.ReplaceAll(cmd, "{PID_FILE_ESC}", strings.ReplaceAll(s.PidPath(), "/", "\\/"))
	cmd = strings.ReplaceAll(cmd, "{APP_PATH}", GetDir(AppDir))
	cmd = strings.ReplaceAll(cmd, "{CONF_FILE}", s.ConfigPath())
	cmd = strings.ReplaceAll(cmd, "{CONF_PATH}", GetDir(ConfDir))
	cmd = strings.ReplaceAll(cmd, "{DATA_PATH}", s.DataPath())
	cmd = strings.ReplaceAll(cmd, "{LOG_PATH}", GetDir(LogDir))
	cmd = strings.ReplaceAll(cmd, "{HOME_PATH}", GetDir(HomeDir))
	return cmd
}
