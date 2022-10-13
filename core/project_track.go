package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
)

// ProjectTileFile is the name of the project tracking file.
const ProjectTrackFile = "projects.json"

// ProjectTrackNameDelimiter is the delimiter of values in the service tracker names.
const ProjectTrackNameDelimiter = "||"

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
	trackFilePath := filepath.Join(GetDir(UserDir), ProjectTrackFile)
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
	trackFilePath := filepath.Join(GetDir(UserDir), ProjectTrackFile)
	if err := ioutil.WriteFile(trackFilePath, rawData, mkdirPerm); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func projectTrackGetServiceName(d interface{}, p *Project) string {
	if d == nil || p == nil {
		return ""
	}
	switch d := d.(type) {
	case def.Service:
		{
			return fmt.Sprintf("s%s%s%s%s%s%s", ProjectTrackNameDelimiter, p.Name, ProjectTrackNameDelimiter, d.Name, ProjectTrackNameDelimiter, d.Type)
		}
	case *def.App:
		{
			return fmt.Sprintf("a%s%s%s%s%s%s", ProjectTrackNameDelimiter, p.Name, ProjectTrackNameDelimiter, d.Name, ProjectTrackNameDelimiter, d.Type)
		}
	}
	return ""
}

// ProjectTrackGetService returns a service from a project track service name.
func ProjectTrackGetService(name string) (Service, error) {
	serviceList, err := LoadServiceList()
	if err != nil {
		return Service{}, err
	}
	values := strings.Split(name, ProjectTrackNameDelimiter)
	if len(values) < 4 {
		return Service{}, ErrInvalidService
	}
	service, err := serviceList.Match(values[3])
	if err != nil {
		return Service{}, err
	}
	mockProject := &Project{Name: values[1]}
	var mockDef interface{}
	switch values[0] {
	case "s":
		{
			mockDef = &def.Service{Name: values[2], Type: values[3]}
			break
		}
	case "a":
		{
			mockDef = &def.App{Name: values[2], Type: values[3]}
		}
	}
	service.SetDefinition(mockProject, mockDef)
	return service, nil
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
	serviceNames := make([]string, 0)
	for _, service := range p.Services {
		serviceNames = append(serviceNames, projectTrackGetServiceName(service, p))
	}
	for _, service := range p.Apps {
		serviceNames = append(serviceNames, projectTrackGetServiceName(service, p))
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
