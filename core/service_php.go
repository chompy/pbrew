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
func (p PHPExtensions) Match(name string) (string, string, error) {
	for extName, extCmd := range p {
		if extName == name {
			return extName, extCmd, nil
		}
	}
	for extName, extCmd := range p {
		if wildcardCompare(extName, name) {
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

// PHPInstallExtension installs the given PHP extension.
func (s *Service) PHPInstallExtension(name string) error {
	phpExtList, err := LoadPHPExtensionList()
	if err != nil {
		return errors.WithStack(err)
	}
	extKey, extCmd, err := phpExtList.Match(name)
	if err != nil {
		return errors.WithStack(errors.WithMessage(err, name))
	}
	done := output.Duration(fmt.Sprintf("Installing PHP extension %s.", extKey))
	cmd := NewShellCommand()
	cmd.Args = []string{"--login", "-c", s.injectCommandParams(extCmd)}
	if err := cmd.Interactive(); err != nil {
		return errors.WithStack(errors.WithMessage(err, extKey))
	}
	done()
	return nil
}

func (s *Service) phpFpmPoolPath() (string, error) {
	if !s.IsPHP() {
		return "", errors.WithStack(errors.WithMessage(ErrInvalidService, s.BrewName))
	}
	verSplit := strings.Split(s.BrewName, "@")
	if len(verSplit) < 2 {
		return "", errors.WithStack(errors.WithMessage(ErrInvalidService, s.BrewName))
	}
	return filepath.Join(BrewPath(), "etc", "php", verSplit[1], "php-fpm.d"), nil
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
	fpmPoolPath, err := s.phpFpmPoolPath()
	if err != nil {
		return errors.WithStack(err)
	}
	if err := ioutil.WriteFile(filepath.Join(fpmPoolPath, fmt.Sprintf("%s_%s.conf", p.Name, d.Name)), []byte(fpmPoolConf), 0755); err != nil {
		return errors.WithStack(err)
	}
	done()
	return nil
}

func (s *Service) phpCleanup(d *def.App, p *Project) error {
	fpmPoolPath, err := s.phpFpmPoolPath()
	if err != nil {
		return errors.WithStack(err)
	}
	os.Remove(filepath.Join(fpmPoolPath, fmt.Sprintf("%s_%s.conf", p.Name, d.Name)))
	return nil
}
