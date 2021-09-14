package core

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"github.com/pkg/errors"
)

type serviceTemplateVars struct {
	Name      string
	Port      int
	Socket    string
	Pid       string
	ConfigDir string
	DataDir   string
	BrewDir   string
	LogDir    string
	User      string
	Group     string
	Params    map[string]interface{}
}

// BuildConfigTemplateVars returns template variables for service config generation.
func (s *Service) BuildConfigTemplateVars() (serviceTemplateVars, error) {
	port, err := s.Port()
	if err != nil {
		return serviceTemplateVars{}, err
	}
	currentUser, err := user.Current()
	if err != nil {
		return serviceTemplateVars{}, errors.WithStack(err)
	}
	currentUserGroup, err := user.LookupGroupId(currentUser.Gid)
	if err != nil {
		return serviceTemplateVars{}, errors.WithStack(err)
	}
	return serviceTemplateVars{
		Name:      strings.ReplaceAll(s.BrewName, "@", "-"),
		Port:      port,
		Socket:    s.SocketPath(),
		Pid:       s.PidPath(),
		ConfigDir: filepath.Dir(s.ConfigPath()),
		DataDir:   s.DataPath(),
		BrewDir:   GetDir(BrewDir),
		LogDir:    GetDir(LogDir),
		User:      currentUser.Username,
		Group:     currentUserGroup.Name,
		Params:    s.ConfigParams(),
	}, nil
}

// GenerateConfigFile generates base config file for service.
func (s *Service) GenerateConfigFile() error {
	done := output.Duration(fmt.Sprintf("Generate config for %s.", s.BrewName))
	// assume config not needed if template not defined
	if s.ConfigTemplates == nil {
		return nil
	}
	// build template vars
	templateVars, err := s.BuildConfigTemplateVars()
	if err != nil {
		return err
	}
	// itterate templates
	for templateFilename, configPath := range s.ConfigTemplates {
		// get paths
		templatePath := filepath.Join(GetDir(AppDir), "conf", templateFilename)
		configPath = s.injectCommandParams(configPath)
		// generate config
		tmpl, err := template.ParseFiles(templatePath)
		if err != nil {
			return errors.WithStack(err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, templateVars); err != nil {
			return errors.WithStack(err)
		}
		// save
		if err := ioutil.WriteFile(configPath, buf.Bytes(), mkdirPerm); err != nil {
			return errors.WithStack(err)
		}
	}
	done()
	return nil

}
