package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

var appYamlFilenames = []string{".platform.app.yaml", ".platform.app.pcc.yaml"}
var serviceYamlFilenames = []string{".platform/services.yaml", ".platform/services.pcc.yaml"}
var routesYamlFilenames = []string{".platform/routes.yaml", ".platform/routes.pcc.yaml"}

// Project defines a Platform.sh project.
type Project struct {
	Path     string
	Name     string
	Apps     []*def.App
	Services []def.Service
	Routes   []def.Route
}

// LoadProject loads a project at given the path.
func LoadProject(projPath string) (*Project, error) {
	var err error
	projPath, err = filepath.Abs(projPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	done := output.Duration(fmt.Sprintf("Read project at %s.", projPath))
	appPaths := scanPlatformAppYaml(projPath, false)
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
	routes, err = def.ExpandRoutes(routes, "platform.cc")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	output.Info(
		fmt.Sprintf("Found %d app(s), %d service(s), %d route(s).", len(apps), len(services), len(routes)),
	)
	done()
	return &Project{
		Path:     projPath,
		Name:     filepath.Base(projPath),
		Apps:     apps,
		Services: services,
		Routes:   routes,
	}, nil
}

// GetBrewServices returns list of Homebrew services used by project.
func (p *Project) GetBrewServices() ([]*Service, error) {
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
	done := output.Duration("Installing services. (THIS MIGHT TAKE A WHILE.)")
	services, err := p.GetBrewServices()
	if err != nil {
		return errors.WithStack(err)
	}
	for _, service := range services {
		if service.IsInstalled() {
			output.LogInfo(fmt.Sprintf("Service '%s' already installed.", service.BrewName))
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

// SetupServices configures services for project.
func (p *Project) SetupServices() error {
	done := output.Duration("Setup services.")
	serviceList, err := LoadServiceList()
	if err != nil {
		return errors.WithStack(err)
	}
	for _, service := range p.Services {
		brewService, err := serviceList.MatchDef(service)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				continue
			}
			return errors.WithStack(err)
		}
		if err := brewService.Setup(&service, p); err != nil {
			return errors.WithStack(err)
		}
	}
	done()
	return nil
}

// Start starts the project.
func (p *Project) Start() error {
	done := output.Duration("Starting services.")
	services, err := p.GetBrewServices()
	if err != nil {
		return errors.WithStack(err)
	}
	// check if brew is installed
	if !IsBrewInstalled() {
		return errors.WithStack(BrewInstall())
	}
	// install services
	if err := p.InstallServices(); err != nil {
		return errors.WithStack(err)
	}
	// start services
	for _, service := range services {
		if err := service.Start(); err != nil {
			return errors.WithStack(err)
		}
	}
	// setup services
	if err := p.SetupServices(); err != nil {
		return errors.WithStack(err)
	}
	done()
	return nil
}

// Stop stops the project.
func (p *Project) Stop() error {
	done := output.Duration("Stopping services.")
	services, err := p.GetBrewServices()
	if err != nil {
		return errors.WithStack(err)
	}
	// stop services
	for _, service := range services {
		if !service.IsRunning() {
			continue
		}
		if err := service.Stop(); err != nil {
			return errors.WithStack(err)
		}
	}
	done()
	return nil
}

// Shell opens a shell for given app.
func (p *Project) Shell(d *def.App) error {
	output.Info(fmt.Sprintf("Access shell for %s.", d.Name))
	// get app brew service
	serviceList, err := LoadServiceList()
	if err != nil {
		return errors.WithStack(err)
	}
	brewAppService, err := serviceList.MatchDef(d)
	if err != nil {
		return errors.WithStack(err)
	}
	if !brewAppService.IsRunning() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotRunning, brewAppService.BrewName))
	}
	// build path
	envPath := make([]string, 0)
	envPath = append(envPath, filepath.Join(BrewPath(), "bin"))
	for _, service := range p.Services {
		brewService, err := serviceList.MatchDef(&service)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				continue
			}
			return errors.WithStack(err)
		}
		if !brewService.IsRunning() {
			return errors.WithStack(errors.WithMessage(ErrServiceNotRunning, brewService.BrewName))
		}
		envPath = append(envPath, filepath.Join(BrewPath(), "opt", brewService.BrewName, "bin"))
	}
	envPath = append(envPath, filepath.Join(BrewPath(), "opt", brewAppService.BrewName, "bin"))
	envPath = append(envPath, "/bin")
	envPath = append(envPath, "/usr/bin")

	// inject env vars
	env := make([]string, 0)
	env = append(env, "PATH="+strings.Join(envPath, ":"))
	env = append(env, fmt.Sprintf("PS1=%s-%s> ", p.Name, d.Name))
	for k, v := range p.Env(d) {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	// build command attributes and run
	cmd := exec.Command("bash", "--norc")
	cmd.Env = env
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return errors.WithStack(errors.WithMessage(err, d.Name))
	}
	return nil
}
