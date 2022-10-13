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
	Name            string            `yaml:"name"`
	BrewName        string            `yaml:"brew_name"`
	PostInstallCmd  string            `yaml:"post_install"`
	PreInstallCmd   string            `yaml:"pre_install"`
	ConfigTemplates map[string]string `yaml:"config_templates"`
	StartCmd        string            `yaml:"start"`
	StopCmd         string            `yaml:"stop"`
	ReloadCmd       string            `yaml:"reload"`
	InstallCheckCmd string            `yaml:"install_check"`
	Dependencies    []string          `yaml:"dependencies"`
	Multiple        bool              `yaml:"multiple"`
	PortOverride    int               `yaml:"port"`
	ForceBottle     bool              `yaml:"force_bottle"`
	ProjectName     string
	usePbrewBottles bool
	project         *Project
	definition      interface{}
}

func (s Service) Empty() bool {
	return s.StartCmd == ""
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
	if s.IsVarnish() && s.project != nil && !s.project.config.EnableVarnish {
		output.Info("Skipping install. ('enable_varnish' is false).")
		return nil
	}
	if s.BrewName != "" {
		installName := s.BrewName
		if s.usePbrewBottles {
			if err := brewBottleDownload(s.BrewName); err == nil {
				installName = brewBottleDownloadPath(s.BrewName)
			}
		}
		forceBottleStr := ""
		if s.ForceBottle {
			forceBottleStr = "--force-bottle"
		}
		if err := brewCommand("install", installName, forceBottleStr); err != nil {
			if !s.InstallCheck() {
				return err
			}
			output.Warn("Install command errored but install check was successful... " + err.Error())
		}
		brewCommand("services", "stop", s.BrewName)
	}
	if err := s.PostInstall(); err != nil {
		return err
	}
	return nil
}

// Uninstall uninstalls the service.
func (s *Service) Uninstall() error {
	if !s.IsInstalled() {
		return nil
	}
	if s.BrewName != "" {
		if err := brewCommand("uninstall", s.BrewName); err != nil {
			return err
		}
	}
	return nil
}

// InstallCheck checks to see if the installation was successful.
func (s Service) InstallCheck() bool {
	// run cmd
	if s.InstallCheckCmd != "" {
		cmdStr := s.injectCommandParams(s.InstallCheckCmd)
		cmd := NewShellCommand()
		cmd.Args = []string{"-c", cmdStr}
		cmd.Env = ServicesEnv([]Service{s})
		if err := cmd.Interactive(); err != nil {
			return false
		}
	}
	return true
}

// PreInstall runs the pre install command for the service.
func (s Service) PreInstall() error {
	// run cmd
	if s.PreInstallCmd != "" {
		cmdStr := s.injectCommandParams(s.PreInstallCmd)
		cmd := NewShellCommand()
		cmd.Args = []string{"-c", cmdStr}
		cmd.Env = ServicesEnv([]Service{s})
		if err := cmd.Interactive(); err != nil {
			return errors.WithMessage(err, s.DisplayName())
		}
	}
	return nil
}

// PostInstall runs the post install command for the service.
func (s Service) PostInstall() error {
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
		cmd.Args = []string{"-c", cmdStr}
		cmd.Env = ServicesEnv([]Service{s})
		if err := cmd.Interactive(); err != nil {
			return errors.WithMessage(err, s.DisplayName())
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
	if !s.InstallCheck() {
		return false
	}
	if s.BrewName != "" {
		info, err := s.Info()
		if err != nil {
			output.Error(err)
			return false
		}
		return len(info["installed"].([]interface{})) > 0
	}
	return true
}

// IsRunning returns true if service is running.
func (s *Service) IsRunning() bool {
	if s.IsSolr() {
		return s.IsSolrRunning()
	} else if s.IsRedis() {
		return s.IsRedisRunning()
	} else if s.IsVarnish() {
		return s.IsVarnishRunning()
	}
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
func (s Service) Start() error {
	done := output.Duration(fmt.Sprintf("Start %s.", s.DisplayName()))
	// varnish disabled
	if s.IsVarnish() && s.project != nil && !s.project.config.EnableVarnish {
		output.Info("Skipping ('enable_varnish' is false).")
		done()
		return nil
	}
	// check status
	if !s.IsInstalled() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotInstalled, s.DisplayName()))
	}
	if s.IsRunning() {
		return errors.WithStack(errors.WithMessage(ErrServiceAlreadyRunning, s.DisplayName()))
	}
	// execute start cmd
	done2 := output.Duration("Start up.")
	cmdStr := s.injectCommandParams(s.StartCmd)
	cmd := NewShellCommand()
	cmd.Env = ServicesEnv([]Service{s})
	if s.project != nil && s.definition != nil {
		switch d := s.definition.(type) {
		case *def.App:
			{
				var err error
				cmd, err = s.project.getAppShellCommand(d)
				if err != nil {
					return err
				}
				break
			}
		}
	}
	cmd.Args = []string{"-c", cmdStr}
	if err := cmd.Interactive(); err != nil {
		return errors.WithMessage(err, s.DisplayName())
	}
	done2()
	done()
	return nil
}

// Stop will stop the service.
func (s Service) Stop() error {
	done := output.Duration(fmt.Sprintf("Stop %s.", s.DisplayName()))
	// check status
	if !s.IsInstalled() {
		return errors.WithStack(ErrServiceNotInstalled)
	}
	if !s.IsRunning() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotRunning, s.DisplayName()))
	}
	// execute stop cmd
	cmdStr := s.injectCommandParams(s.StopCmd)
	cmd := NewShellCommand()
	cmd.Args = []string{"-c", cmdStr}
	cmd.Env = ServicesEnv([]Service{s})
	if err := cmd.Interactive(); err != nil {
		return errors.WithMessage(err, s.DisplayName())
	}
	done()
	return nil
}

// Reload reloads the service configuration.
func (s Service) Reload() error {
	done := output.Duration(fmt.Sprintf("Reload %s.", s.DisplayName()))
	// varnish disabled
	if s.IsVarnish() && s.project != nil && !s.project.config.EnableVarnish {
		output.Info("Skipping ('enable_varnish' is false).")
		done()
		return nil
	}
	// check status
	if s.ReloadCmd == "" {
		return errors.WithStack(errors.WithMessage(ErrServiceReloadNotDefined, s.DisplayName()))
	}
	if !s.IsInstalled() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotInstalled, s.DisplayName()))
	}
	if !s.IsRunning() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotRunning, s.DisplayName()))
	}
	// execute reload cmd
	cmdStr := s.injectCommandParams(s.ReloadCmd)
	cmd := NewShellCommand()
	cmd.Args = []string{"-c", cmdStr}
	cmd.Env = ServicesEnv([]Service{s})
	if err := cmd.Interactive(); err != nil {
		return errors.WithMessage(err, s.DisplayName())
	}
	done()
	return nil
}

// PreStart performs setup that should occur prior to starting service.
func (s *Service) PreStart() error {
	done := output.Duration(fmt.Sprintf("Pre setup %s.", s.DisplayName()))
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
	if s.IsPHP() {
		if err := s.phpPreSetup(); err != nil {
			return err
		}
	}
	done()
	return nil
}

// PostStart performs setup that should occur after starting service.
func (s *Service) PostStart() error {
	// varnish disabled
	if s.IsVarnish() && s.project != nil && !s.project.config.EnableVarnish {
		return nil
	}
	switch d := s.definition.(type) {
	case *def.Service:
		{
			done := output.Duration(fmt.Sprintf("Post setup %s.", d.Name))
			if !s.IsRunning() {
				return errors.WithStack(errors.WithMessage(ErrServiceNotRunning, s.DisplayName()))
			}
			if s.IsMySQL() {
				if err := s.mySQLPostSetup(); err != nil {
					return err
				}
			} else if s.IsSolr() {
				if err := s.solrPostSetup(); err != nil {
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
func (s *Service) Cleanup() error {
	switch d := s.definition.(type) {
	case *def.App:
		{
			done := output.Duration(fmt.Sprintf("Clean up %s.", d.Name))
			if s.IsPHP() {
				if err := s.phpCleanup(); err != nil {
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
func (s *Service) Purge() error {
	switch d := s.definition.(type) {
	case *def.Service:
		{
			done := output.Duration(fmt.Sprintf("Purge %s.", d.Name))
			if s.IsMySQL() {
				if err := s.mySQLPurge(); err != nil {
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
func (s Service) Port() (int, error) {
	if s.PortOverride > 0 {
		return s.PortOverride, nil
	}
	portMap, err := LoadPortMap()
	if err != nil {
		return 0, err
	}
	return portMap.ServicePort(s)
}

// SocketPath returns path to service socket.
func (s *Service) SocketPath() string {
	return filepath.Join(GetDir(RunDir), fmt.Sprintf("%s.sock", s.UniqueName()))
}

// UpstreamSocketPath returns path to app upstream socket.
func (s *Service) UpstreamSocketPath() string {
	return s.SocketPath()
}

// BrewAppName returns the brew app name without namespace.
func (s *Service) BrewAppName() string {
	return brewAppName(s.BrewName)
}

// DisplayName returns the name the service should be displayed to the user as.
func (s *Service) DisplayName() string {
	if s.Name != "" {
		return s.Name
	}
	return s.BrewAppName()
}

// UniqueName returns a unique name for this instance.
func (s *Service) UniqueName() string {
	brewName := strings.ReplaceAll(s.BrewAppName(), "@", "-")
	if s.Multiple && s.project != nil && s.definition != nil {
		switch d := s.definition.(type) {
		case *def.App:
			{
				return fmt.Sprintf("%s-%s-%s", brewName, d.Name, s.project.Name)
			}
		case *def.Service:
			{
				return fmt.Sprintf("%s-%s-%s", brewName, d.Name, s.project.Name)
			}
		default:
			{
				return fmt.Sprintf("%s-%s", brewName, s.project.Name)
			}
		}

	}
	return brewName
}

// PidPath returns path to service pid file.
func (s *Service) PidPath() string {
	return filepath.Join(GetDir(RunDir), fmt.Sprintf("%s.pid", s.UniqueName()))
}

// ConfigPath returns path to service config file.
func (s *Service) ConfigPath() string {
	return filepath.Join(GetDir(ConfDir), fmt.Sprintf("%s.conf", s.UniqueName()))
}

// DataPath returns path to service data directory.
func (s *Service) DataPath() string {
	name := s.BrewAppName()
	if s.BrewName == "" {
		name = strings.ToLower(s.Name)
	}
	return filepath.Join(GetDir(DataDir), strings.ReplaceAll(name, "@", "-"))
}

// ConfigParams returns confir template parameters for service.
func (s *Service) ConfigParams() map[string]interface{} {
	if s.IsPHP() {
		return s.phpConfigParams()
	}
	if s.IsVarnish() {
		return s.varnishConfigParams()
	}
	return map[string]interface{}{}
}

// SetDefinition set the project and service definition for this service.
func (s *Service) SetDefinition(p *Project, d interface{}) {
	s.project = p
	s.definition = d
}

func (s *Service) injectCommandParams(cmd string) string {
	port, err := s.Port()
	if err != nil {
		output.Warn(err.Error())
	}
	cmd = strings.ReplaceAll(cmd, "{BREW_PATH}", GetDir(BrewDir))
	cmd = strings.ReplaceAll(cmd, "{BREW_APP}", s.BrewAppName())
	cmd = strings.ReplaceAll(cmd, "{NAME}", s.Name)
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
