package core

import (
	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
)

// ServiceOverride defines a custom override for a service.
type ServiceOverride struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Port     int    `yaml:"port"`
	Host     string `yaml:"host"`
	Path     string `yaml:"path"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Scheme   string `yaml:"scheme"`
}

func (s ServiceOverride) Relationship() map[string]interface{} {
	return map[string]interface{}{
		"host":        s.Host,
		"hostname":    s.Host,
		"ip":          s.Host,
		"port":        s.Port,
		"username":    s.Username,
		"password":    s.Password,
		"path":        s.Path,
		"scheme":      s.Scheme,
		"host_mapped": false,
		"pubilc":      false,
		"query": map[string]interface{}{
			"is_master": true,
		},
	}
}

// MatchServiceOverride matches name with a service override.
func MatchServiceOverride(name string) (*ServiceOverride, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	for i := range config.ServiceOverrides {
		if config.ServiceOverrides[i].Type == name {
			return &config.ServiceOverrides[i], nil
		}
	}
	for i := range config.ServiceOverrides {
		if wildcardCompare(name, config.ServiceOverrides[i].Type) {
			return &config.ServiceOverrides[i], nil
		}
	}
	return nil, errors.WithStack(errors.WithMessage(ErrServiceNotFound, name))
}

// MatchServiceOverrideDef matches service definition with a service override.
func MatchServiceOverrideDef(d interface{}) (*ServiceOverride, error) {
	switch d := d.(type) {
	case *def.App:
		{
			service, err := MatchServiceOverride(d.Type)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return service, nil
		}
	case *def.Service:
		{
			service, err := MatchServiceOverride(d.Type)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return service, nil
		}
	case def.Service:
		{
			service, err := MatchServiceOverride(d.Type)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return service, nil
		}
	case *def.AppWorker:
		{
			service, err := MatchServiceOverride(d.Type)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return service, nil
		}
	}
	return nil, errors.WithStack(ErrInvalidDef)
}

// ServiceHasOverride returns true if give service def has override.
func ServiceHasOverride(d interface{}) bool {
	override, _ := MatchServiceOverrideDef(d)
	return override != nil
}
