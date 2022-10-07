package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"

	"github.com/pkg/errors"
)

const portMapFile = "ports.json"

// PortMap maps service to it ports.
type PortMap map[string]int

// LoadPortMap loads the port mappings.
func LoadPortMap() (PortMap, error) {
	pathTo := filepath.Join(GetDir(UserDir), portMapFile)
	portJSON, err := ioutil.ReadFile(pathTo)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(PortMap), nil
		}
		return nil, errors.WithStack(err)
	}
	out := make(PortMap)
	if err := json.Unmarshal(portJSON, &out); err != nil {
		return nil, errors.WithStack(err)
	}
	return out, nil
}

// Save stores the port mappings to file.
func (p PortMap) save() error {
	pathTo := filepath.Join(GetDir(UserDir), portMapFile)
	portJSON, err := json.Marshal(p)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := ioutil.WriteFile(pathTo, portJSON, mkdirPerm); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (p PortMap) assignPort(name string) (int, error) {
	if name == "" {
		return 0, errors.WithStack(ErrInvalidDef)
	}
	if p[name] != 0 {
		return p[name], nil
	}
	config, err := LoadConfig()
	if err != nil {
		return 0, err
	}
	// look for free port
	// TODO set max range
	currentPort := config.PortRangeStart
	for {
		isAvailable := true
		for _, port := range p {
			if port == currentPort {
				isAvailable = false
				break
			}
		}
		if isAvailable {
			break
		}
		currentPort++
	}
	// save
	p[name] = currentPort
	if err := p.save(); err != nil {
		return currentPort, errors.WithStack(err)
	}
	return currentPort, nil
}

// ServicePort retrieves or creates an assigned port for the given service.
func (p PortMap) ServicePort(s Service) (int, error) {
	appName := s.BrewAppName()
	if appName == "" {
		appName = s.Name
	}
	if appName == "" {
		return 0, ErrServiceNoName
	}
	if s.Multiple && s.project != nil {
		// multi-instance service
		return p.assignPort("s-" + appName + "-" + s.project.Name)
	}
	return p.assignPort("s-" + s.Name)
}

// UpstreamPort retrieves of creates an assigned port for the given app def.
func (p PortMap) UpstreamPort(a *def.App, proj *Project) (int, error) {
	return p.assignPort(fmt.Sprintf("u-%s-%s", proj.Name, a.Name))
}
