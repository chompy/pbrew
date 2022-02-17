package core

import (
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"

	"github.com/pkg/errors"
)

var loadedServiceList map[string]*Service

// ServiceList is a list of available Homebrew services.
type ServiceList map[string]*Service

// LoadServiceList loads all available Homebrew services.
func LoadServiceList() (ServiceList, error) {
	if loadedServiceList != nil {
		return loadedServiceList, nil
	}
	done := output.Duration("Load Homebrew service list.")
	loadedServiceList = make(ServiceList)
	if err := loadYAML("services", loadedServiceList); err != nil {
		return nil, errors.WithStack(err)
	}
	done()
	return loadedServiceList, nil
}

// Match matches platform.sh service with homebrew service.
func (s ServiceList) Match(name string) (*Service, error) {
	matchName := strings.ReplaceAll(name, ":", "-")
	for serviceName, service := range s {
		serviceName = strings.ReplaceAll(serviceName, ":", "-")
		if serviceName == matchName {
			return service, nil
		}
	}
	for serviceName, service := range s {
		serviceName = strings.ReplaceAll(serviceName, ":", "-")
		if wildcardCompare(matchName, serviceName) {
			return service, nil
		}
	}
	return nil, errors.WithStack(errors.WithMessage(ErrServiceNotFound, name))
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
	case def.Service:
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
