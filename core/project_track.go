package core

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

// ProjectTileFile is the name of the project tracking file.
const ProjectTrackFile = "projects.json"

// ProjectTrack tracks running project.
type ProjectTrack struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Services []string  `json:"services"`
	Time     time.Time `json:"time"`
}

var projectTracks []ProjectTrack

func loadProjectTracks() error {
	projectTracks = make([]ProjectTrack, 0)
	workingTracks := make([]ProjectTrack, 0)
	trackFilePath := filepath.Join(GetDir(UserDir), ProjectTrackFile)
	rawData, err := ioutil.ReadFile(trackFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return errors.WithStack(err)
	}
	if err := json.Unmarshal(rawData, &workingTracks); err != nil {
		return errors.WithStack(err)
	}

	sysinfo := syscall.Sysinfo_t{}
	if err := syscall.Sysinfo(&sysinfo); err != nil {
		return errors.WithStack(err)
	}
	bootTime := time.Now().Add(time.Second * -time.Duration(sysinfo.Uptime))
	for _, proj := range workingTracks {
		if proj.Time.Before(bootTime) {
			continue
		}
		projectTracks = append(projectTracks, proj)
	}
	return nil
}

func saveProjectTracks() error {
	rawData, err := json.Marshal(projectTracks)
	if err != nil {
		return errors.WithStack(err)
	}
	trackFilePath := filepath.Join(GetDir(UserDir), ProjectTrackFile)
	if err := ioutil.WriteFile(trackFilePath, rawData, mkdirPerm); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// ProjectTrackGet returns list of tracked running projects.
func ProjectTrackGet() ([]ProjectTrack, error) {
	if err := loadProjectTracks(); err != nil {
		return nil, err
	}
	return projectTracks, nil
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
		serviceNames = append(serviceNames, service.BrewAppName())
	}
	pt := ProjectTrack{
		Name:     p.Name,
		Path:     p.Path,
		Services: serviceNames,
		Time:     time.Now(),
	}
	if err := loadProjectTracks(); err != nil {
		return err
	}
	for _, pt := range projectTracks {
		if pt.Name == p.Name && pt.Path == p.Path {
			return nil
		}
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
