package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

var appYamlFilenames = []string{".platform.app.yaml", ".platform.app.pcc.yaml"}
var serviceYamlFilenames = []string{".platform/services.yaml", ".platform/services.pcc.yaml"}
var routesYamlFilenames = []string{".platform/routes.yaml", ".platform/routes.pcc.yaml"}

// Project defines a Platform.sh project.
type Project struct {
	Path            string        `json:"path"`
	Name            string        `json:"name"`
	Apps            []*def.App    `json:"-"`
	Services        []def.Service `json:"-"`
	Routes          []def.Route   `json:"-"`
	NoMounts        bool          `json:"-"`
	UsePbrewBottles bool          `json:"-"`
}

func findProjectRoot(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", errors.WithStack(err)
	}
	pathSplit := strings.Split(path, string(os.PathSeparator))
	if pathSplit[0] == "" {
		pathSplit[0] = string(os.PathSeparator)
	}
	scanFiles := []string{".platform.app.yaml", ".platform"}
	for i := range pathSplit {
		currentPath := filepath.Join(pathSplit[0 : i+1]...)
		for _, scanFile := range scanFiles {
			if _, err := os.Stat(filepath.Join(currentPath, scanFile)); err == nil {
				return currentPath, nil
			}
		}
	}
	return "", ErrProjectNotFound
}

// LoadProject loads a project at given the path.
func LoadProject(projPath string) (*Project, error) {
	var err error
	projPath, err = findProjectRoot(projPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	done := output.Duration(fmt.Sprintf("Read project at %s.", projPath))
	appPaths := scanPlatformAppYaml(projPath, false)
	apps := make([]*def.App, 0)
	for _, appPath := range appPaths {
		app, err := def.ParseAppYamlFiles(appPath, nil)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	serviceYamlFullPaths := make([]string, 0)
	for _, f := range serviceYamlFilenames {
		serviceYamlFullPaths = append(serviceYamlFullPaths, filepath.Join(projPath, f))
	}
	services, err := def.ParseServiceYamlFiles(serviceYamlFullPaths)
	if err != nil {
		return nil, err
	}
	routesYamlFullPaths := make([]string, 0)
	for _, f := range routesYamlFilenames {
		routesYamlFullPaths = append(routesYamlFullPaths, filepath.Join(projPath, f))
	}
	routes, err := def.ParseRoutesYamlFiles(routesYamlFullPaths)
	if err != nil {
		return nil, err
	}
	routes, err = def.ExpandRoutes(routes, "platform.cc")
	if err != nil {
		return nil, err
	}
	output.Info(
		fmt.Sprintf("Found %d app(s), %d service(s), %d route(s).", len(apps), len(services), len(routes)),
	)
	done()
	return &Project{
		Path:     projPath,
		Name:     strings.ToLower(filepath.Base(projPath)),
		Apps:     apps,
		Services: services,
		Routes:   routes,
		NoMounts: false,
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
		if ServiceHasOverride(app) {
			continue
		}
		service, err := serviceList.Match(app.Type)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				continue
			}
			return nil, errors.WithStack(err)
		}
		service.project = p
		service.definition = app
		out = append(out, service)
	}
	for i, pshs := range p.Services {
		if ServiceHasOverride(pshs) {
			continue
		}
		service, err := serviceList.Match(pshs.Type)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				continue
			}
			return nil, errors.WithStack(err)
		}
		service.project = p
		service.definition = &p.Services[i]
		out = append(out, service)
	}
	return out, nil
}

// InstallServices installs all Homebrew services for project.
func (p *Project) InstallServices() error {
	done := output.Duration("Installing services. (THIS MIGHT TAKE A WHILE.)")
	services, err := p.GetBrewServices()
	if err != nil {
		return err
	}
	for _, service := range services {
		service.usePbrewBottles = p.UsePbrewBottles
		if err := service.InstallDependencies(); err != nil {
			return err
		}
		if service.IsInstalled() {
			output.LogInfo(fmt.Sprintf("Service '%s' already installed.", service.DisplayName()))
			continue
		}
		d2 := output.Duration(fmt.Sprintf("Installing '%s' service.", service.DisplayName()))
		if err := service.Install(); err != nil {
			return err
		}
		d2()
	}
	done()
	return nil
}

// PreSetup configures services for project.
func (p *Project) PreSetup() error {
	done := output.Duration("Services pre-setup.")
	serviceList, err := LoadServiceList()
	if err != nil {
		return err
	}
	servicePreSetup := func(service interface{}) error {
		if ServiceHasOverride(service) {
			return nil
		}
		brewService, err := serviceList.MatchDef(service)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				return nil
			}
			return err
		}
		brewService.project = p
		brewService.definition = service
		if err := brewService.PreStart(); err != nil {
			return err
		}
		return nil
	}
	for _, service := range p.Services {
		if err := servicePreSetup(&service); err != nil {
			return err
		}
	}
	for _, service := range p.Apps {
		if err := servicePreSetup(service); err != nil {
			return err
		}
	}
	done()
	return nil
}

// PostSetup configures services for project, post start.
func (p *Project) PostSetup() error {
	done := output.Duration("Services post-setup.")
	serviceList, err := LoadServiceList()
	if err != nil {
		return err
	}
	for _, service := range p.Services {
		if ServiceHasOverride(service) {
			return nil
		}
		brewService, err := serviceList.MatchDef(service)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				continue
			}
			return err
		}
		brewService.project = p
		brewService.definition = &service
		if err := brewService.PostStart(); err != nil {
			return err
		}
	}
	for _, service := range p.Apps {
		brewService, err := serviceList.MatchDef(service)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				continue
			}
			return err
		}
		brewService.project = p
		brewService.definition = service
		if err := brewService.PostStart(); err != nil {
			return err
		}
	}
	done()
	return nil
}

// Start starts the project.
func (p *Project) Start() error {
	done := output.Duration("Starting services.")
	// check if brew is installed
	if !IsBrewInstalled() {
		if err := BrewInstall(); err != nil {
			return err
		}
	}
	// install services
	if err := p.InstallServices(); err != nil {
		return err
	}
	// setup mount symlinks
	if err := p.SetupMounts(); err != nil {
		return err
	}
	// pre-setup services
	if err := p.PreSetup(); err != nil {
		return err
	}
	// start services
	services, err := p.GetBrewServices()
	if err != nil {
		return err
	}
	for _, service := range services {
		// reload if already running
		if service.IsRunning() {
			if err := service.Reload(); err != nil {
				if errors.Is(err, ErrServiceReloadNotDefined) {
					output.Warn(err.Error())
					output.IndentLevel--
					continue
				}
				return err
			}
			continue
		}
		// start
		if err := service.Start(); err != nil {
			if !errors.Is(err, ErrServiceAlreadyRunning) {
				output.Warn(err.Error())
				output.IndentLevel--
				return err
			}
		}
	}
	// setup services
	time.Sleep(time.Second * 2)
	if err := p.PostSetup(); err != nil {
		return err
	}
	done()
	// track project
	if err := ProjectTrackAdd(p); err != nil {
		return err
	}
	return nil
}

// Stop stops the project.
func (p *Project) Stop() error {
	done := output.Duration("Stopping services.")
	// remove from project tracking
	if err := ProjectTrackRemove(p); err != nil {
		return err
	}
	remainingServices, err := ProjectTrackServices()
	if err != nil {
		return err
	}
	// stop services that are no longer needed
	brewServiceList, err := LoadServiceList()
	if err != nil {
		return err
	}
	stopService := func(service interface{}) error {
		if ServiceHasOverride(service) {
			return nil
		}
		brewService, err := brewServiceList.MatchDef(service)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				return nil
			}
			return err
		}
		brewService.project = p
		brewService.definition = service
		if !brewService.IsRunning() {
			return nil
		}
		if err := brewService.Cleanup(); err != nil {
			return err
		}
		// if service is used by other projects then don't stop it, reload instead
		if !brewService.Multiple {
			for _, runningService := range remainingServices {
				if runningService == brewService.BrewAppName() {
					if err := brewService.Reload(); err != nil {
						if errors.Is(err, ErrServiceReloadNotDefined) {
							output.IndentLevel--
							output.Warn(err.Error())
							return nil
						}
						return err
					}
					return nil
				}
			}
		}
		// stop service when no longer needed
		if err := brewService.Stop(); err != nil {
			if !errors.Is(err, ErrServiceNotRunning) {
				return err
			}
			output.IndentLevel--
			output.Warn(err.Error())
		}
		return nil
	}
	for _, service := range p.Services {
		if err := stopService(&service); err != nil {
			return err
		}
	}
	for _, service := range p.Apps {
		if err := stopService(service); err != nil {
			return err
		}
	}
	done()
	return nil
}

// Purge deletes all files for this project created by pbrew.
func (p *Project) Purge() error {
	// stop project
	if err := p.Stop(); err != nil {
		return err
	}
	// purge services
	done := output.Duration("Purging service data.")
	brewServiceList, err := LoadServiceList()
	if err != nil {
		return err
	}
	purgeService := func(service interface{}) error {
		brewService, err := brewServiceList.MatchDef(service)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				return nil
			}
			return err
		}
		brewService.project = p
		brewService.definition = service
		if err := brewService.Purge(); err != nil {
			return err
		}
		return nil
	}
	for _, service := range p.Services {
		if err := purgeService(&service); err != nil {
			return err
		}
	}
	for _, service := range p.Apps {
		if err := purgeService(service); err != nil {
			return err
		}
	}
	done()
	done = output.Duration("Delete data.")
	// delete mnt
	os.RemoveAll(filepath.Join(GetDir(MntDir), p.Name))
	// delete var
	os.Remove(variablePath(p.Name))
	done()
	return nil
}
