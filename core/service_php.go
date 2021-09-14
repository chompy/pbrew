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
	return strings.HasPrefix(s.BrewName, "php")
}

// PHPVersion returns the PHP version.
func (s *Service) PHPVersion() string {
	nameSplit := strings.Split(s.BrewName, "@")
	if len(nameSplit) < 2 {
		return ""
	}
	return strings.TrimSpace(nameSplit[1])
}

// PHPInstallExtension installs the given PHP extension.
func (s *Service) PHPInstallExtension(name string) error {
	if !s.IsPHP() {
		return errors.WithStack(errors.WithMessage(ErrInvalidService, s.BrewName))
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
	cmd.Args = []string{"--login", "-c", s.injectCommandParams(extCmd)}
	cmd.Env = ServicesEnv([]*Service{s})
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

func (s *Service) phpFpmPoolPath(d *def.App, p *Project) string {
	brewName := strings.ReplaceAll(s.BrewName, "@", "-")
	return filepath.Join(GetDir(ConfDir), fmt.Sprintf("%s_%s_%s.conf", brewName, p.Name, d.Name))
}

func (s *Service) phpPreSetup(d *def.App, p *Project) error {
	// checks
	if !s.IsPHP() {
		return errors.WithStack(errors.WithMessage(ErrInvalidService, s.BrewName))
	}
	if s.IsRunning() {
		return errors.WithStack(errors.WithMessage(ErrServiceAlreadyRunning, s.BrewName))
	}
	// install extensions
	for _, ext := range d.Runtime.Extensions {
		if err := s.PHPInstallExtension(ext.Name); err != nil {
			if errors.Is(err, ErrPHPExtNotFound) {
				continue
			}
			return errors.WithStack(err)
		}
	}
	// php fpm pool
	done := output.Duration("Generate PHP FPM pool.")
	fpmPoolConf, err := p.GeneratePhpFpmPool(d)
	if err != nil {
		return errors.WithStack(errors.WithMessage(err, s.BrewName))
	}
	if err := ioutil.WriteFile(s.phpFpmPoolPath(d, p), []byte(fpmPoolConf), 0755); err != nil {
		return errors.WithStack(err)
	}
	done()
	return nil
}

func (s *Service) phpCleanup(d *def.App, p *Project) error {
	os.Remove(filepath.Join(s.phpFpmPoolPath(d, p)))
	return nil
}

func (s *Service) phpConfigParams() map[string]interface{} {
	return map[string]interface{}{
		"Extensions": s.PHPGetInstalledExtensions(),
	}
}
