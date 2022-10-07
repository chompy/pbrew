package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

var loadedPHPExtensionList PHPExtensions

// PHPExtensions is a map of PHP extension name to the command needed to install it.
type PHPExtensions map[string]string

// Match finds a PHP extension that matches given name.
func (p PHPExtensions) Match(name string, version string) (string, string, error) {
	versionName := fmt.Sprintf("%s-%s", version, name)
	// match with php version
	for extName, extCmd := range p {
		if extName == versionName {
			return extName, extCmd, nil
		}
	}
	// match with just extension name
	for extName, extCmd := range p {
		if extName == name {
			return extName, extCmd, nil
		}
	}
	return "", "", errors.WithStack(errors.WithMessage(ErrPHPExtNotFound, name))
}

// LoadPHPExtensionList loads list of PHP extensions.
func LoadPHPExtensionList() (PHPExtensions, error) {
	if loadedPHPExtensionList != nil {
		return loadedPHPExtensionList, nil
	}
	done := output.Duration("Load PHP extension list.")
	loadedPHPExtensionList = make(PHPExtensions)
	if err := loadYAML("php_ext", loadedPHPExtensionList); err != nil {
		return nil, errors.WithStack(err)
	}
	done()
	return loadedPHPExtensionList, nil
}

// IsPHP returns true if service is php.
func (s *Service) IsPHP() bool {
	return strings.HasPrefix(s.BrewAppName(), "php")
}

// PHPVersion returns the PHP version.
func (s *Service) PHPVersion() string {
	nameSplit := strings.Split(s.BrewAppName(), "@")
	if len(nameSplit) < 2 {
		return ""
	}
	return strings.TrimSpace(nameSplit[1])
}

// PHPInstallExtension installs the given PHP extension.
func (s Service) PHPInstallExtension(name string) error {
	if !s.IsPHP() {
		return errors.WithStack(errors.WithMessage(ErrInvalidService, s.DisplayName()))
	}
	phpExtList, err := LoadPHPExtensionList()
	if err != nil {
		return errors.WithStack(err)
	}
	extKey, extCmd, err := phpExtList.Match(name, s.PHPVersion())
	if err != nil {
		return errors.WithStack(errors.WithMessage(err, name))
	}
	done := output.Duration(fmt.Sprintf("Installing PHP extension %s.", extKey))
	cmd := NewShellCommand()
	cmd.Args = []string{"-c", s.injectCommandParams(extCmd)}
	cmd.Env = ServicesEnv([]Service{s})
	if err := cmd.Interactive(); err != nil {
		return errors.WithStack(errors.WithMessage(err, extKey))
	}
	done()
	return nil
}

// PHPGetInstalledExtensions returns list of installed PHP extensions.
func (s *Service) PHPGetInstalledExtensions() []string {
	if !s.IsPHP() {
		return []string{}
	}
	fileInfo, err := ioutil.ReadDir(s.DataPath())
	if err != nil {
		output.Warn(err.Error())
		return []string{}
	}
	out := make([]string, 0)
	for _, file := range fileInfo {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) == ".so" {
			out = append(out, strings.Split(file.Name(), ".")[0])
		}
	}
	return out
}

func (s *Service) phpFpmPoolPath() string {
	brewName := strings.ReplaceAll(s.BrewAppName(), "@", "-")
	switch d := s.definition.(type) {
	case *def.App:
		{
			return filepath.Join(GetDir(ConfDir), fmt.Sprintf("%s_%s_%s.conf", brewName, s.project.Name, d.Name))
		}
	}
	return filepath.Join(GetDir(ConfDir), fmt.Sprintf("%s.conf", brewName))
}

func (s *Service) phpPreSetup() error {
	// checks
	if !s.IsPHP() {
		return errors.WithStack(errors.WithMessage(ErrInvalidService, s.DisplayName()))
	}
	switch d := s.definition.(type) {
	case *def.App:
		{
			// install extensions
			for _, ext := range d.Runtime.Extensions {
				if err := s.PHPInstallExtension(ext.Name); err != nil {
					if errors.Is(err, ErrPHPExtNotFound) {
						continue
					}
					return errors.WithStack(err)
				}
			}
			// (re)generate config file
			// TODO better way??
			if err := s.GenerateConfigFile(); err != nil {
				return err
			}
		}
	default:
		{
			return errors.WithStack(errors.WithMessage(ErrServiceDefNotDefined, s.DisplayName()))
		}
	}
	return nil
}

func (s *Service) phpCleanup() error {
	// delete fpm pool conf
	os.Remove(filepath.Join(s.phpFpmPoolPath()))
	return nil
}

func (s *Service) phpConfigParams() map[string]interface{} {
	vars := def.Variables{}
	projName := ""
	if s.project != nil && s.definition != nil {
		vars, _ = s.project.Variables(s.definition)
		projName = s.project.Name
	}
	return map[string]interface{}{
		"Extensions": s.PHPGetInstalledExtensions(),
		"Ini":        vars.GetStringSubMap("php"),
		"Project":    projName,
	}
}
