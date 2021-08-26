package core

import (
	"fmt"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

var appYamlFilenames = []string{".platform.app.yaml", ".platform.app.pcc.yaml"}
var serviceYamlFilenames = []string{".platform/services.yaml", ".platform/services.pcc.yaml"}
var routesYamlFilenames = []string{".platform/routes.yaml", ".platform/routes.pcc.yaml"}

// Project defines a Platform.sh project.
type Project struct {
	Apps     []*def.App
	Services []def.Service
	Routes   []def.Route
}

// LoadProject loads a project at given the path.
func LoadProject(path string) (*Project, error) {
	done := output.Duration(fmt.Sprintf("Read project at %s.", path))
	appPaths := scanPlatformAppYaml(path, false)
	apps := make([]*def.App, 0)
	for _, appPath := range appPaths {
		app, err := def.ParseAppYamlFiles(appPath, nil)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		apps = append(apps, app)
	}
	services, err := def.ParseServiceYamlFiles(serviceYamlFilenames)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	routes, err := def.ParseRoutesYamlFiles(routesYamlFilenames)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	output.Info(
		fmt.Sprintf("Found %d app(s), %d service(s), %d route(s).", len(apps), len(services), len(routes)),
	)
	done()
	return &Project{
		Apps:     apps,
		Services: services,
		Routes:   routes,
	}, nil
}

// GetServices returns list of Homebrew services used by project.
func (p *Project) GetServices() ([]*Service, error) {
	serviceList, err := LoadServiceList()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	out := make([]*Service, 0)
	for _, app := range p.Apps {
		service, err := serviceList.Match(app.Type)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				continue
			}
			return nil, errors.WithStack(err)
		}
		out = append(out, service)
	}
	for _, pshs := range p.Services {
		service, err := serviceList.Match(pshs.Type)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				continue
			}
			return nil, errors.WithStack(err)
		}
		out = append(out, service)
	}
	return out, nil
}

// InstallServices installs all Homebrew services for project.
func (p *Project) InstallServices() error {
	done := output.Duration("Installing services.")
	services, err := p.GetServices()
	if err != nil {
		return errors.WithStack(err)
	}
	for _, service := range services {
		if service.IsInstalled() {
			output.Info(fmt.Sprintf("Service '%s' already installed.", service.BrewName))
			continue
		}
		d2 := output.Duration(fmt.Sprintf("Installing '%s' service.", service.BrewName))
		if err := service.Install(); err != nil {
			return errors.WithStack(err)
		}
		d2()
	}
	done()
	return nil
}

// Start starts the project.
func (p *Project) Start() error {
	done := output.Duration("Starting services.")
	services, err := p.GetServices()
	if err != nil {
		return errors.WithStack(err)
	}
	// check installed
	for _, service := range services {
		if !service.IsInstalled() {
			return errors.WithStack(errors.WithMessage(ErrServiceNotInstalled, service.BrewName))
		}
	}
	// start services
	for _, service := range services {
		if err := service.Start(); err != nil {
			return errors.WithStack(err)
		}
	}
	done()
	return nil
}
