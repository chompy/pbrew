package core

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const serviceListYAML = "conf/services.yaml"

// ServiceList is a list of available Homebrew services.
type ServiceList map[string]*Service

// LoadServiceList loads all available Homebrew services.
func LoadServiceList() (ServiceList, error) {
	yamlPath, err := os.Executable()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	yamlRaw, err := ioutil.ReadFile(filepath.Join(filepath.Dir(yamlPath), serviceListYAML))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	out := make(ServiceList)
	if err := yaml.Unmarshal(yamlRaw, &out); err != nil {
		return nil, errors.WithStack(err)
	}
	return out, nil
}

// Match matches platform.sh service with homebrew service.
func (s ServiceList) Match(name string) (*Service, error) {
	for serviceName, serviceDef := range s {
		serviceName = strings.ReplaceAll(serviceName, "-", ":")
		if serviceName == name {
			return serviceDef, nil
		}
	}
	for serviceName, serviceDef := range s {
		serviceName = strings.ReplaceAll(serviceName, "-", ":")
		if wildcardCompare(name, serviceName) {
			return serviceDef, nil
		}
	}
	return nil, errors.WithStack(ErrServiceNotFound)
}

// MatchDef matches definition with its homebrew service.
func (s ServiceList) MatchDef(d interface{}) (*Service, error) {
	switch d := d.(type) {
	case *def.App:
		{
			service, err := s.Match(d.Type)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return service, nil
		}
	case *def.Service:
		{
			service, err := s.Match(d.Type)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return service, nil
		}
	case *def.AppWorker:
		{
			service, err := s.Match(d.Type)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return service, nil
		}
	}
	return nil, errors.WithStack(ErrInvalidDef)
}
