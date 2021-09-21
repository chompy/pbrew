package core

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const projectTrackFile = "projects.json"

// ProjectTrack tracks running project.
type ProjectTrack struct {
	Name     string   `json:"name"`
	Path     string   `json:"path"`
	Services []string `json:"services"`
}

var projectTracks []ProjectTrack

func loadProjectTracks() error {
	projectTracks = make([]ProjectTrack, 0)
	trackFilePath := filepath.Join(GetDir(UserDir), projectTrackFile)
	rawData, err := ioutil.ReadFile(trackFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return errors.WithStack(err)
	}
	if err := json.Unmarshal(rawData, &projectTracks); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func saveProjectTracks() error {
	rawData, err := json.Marshal(projectTracks)
	if err != nil {
		return errors.WithStack(err)
	}
	trackFilePath := filepath.Join(GetDir(UserDir), projectTrackFile)
	if err := ioutil.WriteFile(trackFilePath, rawData, mkdirPerm); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// ProjectTrackGet returns list of tracked running projects.
func ProjectTrackGet() []ProjectTrack {
	return projectTracks
}

// ProjectTrackServices returns list of all running services.
func ProjectTrackServices() ([]string, error) {
	if err := loadProjectTracks(); err != nil {
		return nil, err
	}
	out := make([]string, 0)
	for _, pt := range projectTracks {
		hasService := false
		for _, service := range pt.Services {
			for _, addedService := range out {
				if addedService == service {
					hasService = true
					break
				}
			}
			if !hasService {
				out = append(out, service)
			}
		}
	}
	return out, nil
}

// ProjectTrackAdd adds project to tracking.
func ProjectTrackAdd(p *Project) error {
	brewServices, err := p.GetBrewServices()
	if err != nil {
		return err
	}
	serviceNames := make([]string, 0)
	for _, service := range brewServices {
		serviceNames = append(serviceNames, service.BrewName)
	}
	pt := ProjectTrack{
		Name:     p.Name,
		Path:     p.Path,
		Services: serviceNames,
	}
	if err := loadProjectTracks(); err != nil {
		return err
	}
	projectTracks = append(projectTracks, pt)
	if err := saveProjectTracks(); err != nil {
		return err
	}
	return nil
}

// Remove removes project from tracking.
func ProjectTrackRemove(p *Project) error {
	if err := loadProjectTracks(); err != nil {
		return err
	}
	for i, pt := range projectTracks {
		if pt.Name == p.Name && pt.Path == p.Path {
			projectTracks = append(projectTracks[:i], projectTracks[i+1:]...)
			return saveProjectTracks()
		}
	}
	return nil
}
