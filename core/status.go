package core

import (
	"log"
	"sort"
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
)

const serviceStatusNotInstalled = "not installed"
const serviceStatusStopped = "stopped"
const serviceStatusRunning = "running"

// ServiceStatus defines
type ServiceStatus struct {
	DefName      string `json:"def_name"`
	DefType      string `json:"def_type"`
	BrewName     string `json:"brew_name"`
	InstanceName string `json:"instance_name"`
	Project      string `json:"project"`
	Status       string `json:"status"`
	Port         int    `json:"port"`
}

// GetServiceStatuses returns status of all services.
func GetServiceStatuses() ([]ServiceStatus, error) {
	brewServices, err := LoadServiceList()
	if err != nil {
		return nil, err
	}
	projectTracks, err := ProjectTrackGet()
	if err != nil {
		return nil, err
	}
	nginxService := NginxService()
	brewServices[nginxService.BrewAppName()] = nginxService
	// list project statuses
	out := make([]ServiceStatus, 0)
	for _, projTrack := range projectTracks {
		for _, ptServiceName := range projTrack.Services {
			service, err := ProjectTrackGetService(ptServiceName)
			if err != nil {
				log.Println(ptServiceName)
				return nil, err
			}
			defType := ""
			defName := ""
			switch d := service.definition.(type) {
			case *def.App:
				{
					defType = d.Type
					defName = d.Name
					break
				}
			case *def.Service:
				{
					defType = d.Type
					defName = d.Name
				}
			}
			status := serviceStatusNotInstalled
			if service.IsRunning() {
				status = serviceStatusRunning
			} else if service.IsInstalled() && status == serviceStatusNotInstalled {
				status = serviceStatusStopped
			}
			port, err := service.Port()
			if err != nil {
				return nil, err
			}
			out = append(out, ServiceStatus{
				DefName:      defName,
				DefType:      defType,
				BrewName:     service.BrewAppName(),
				InstanceName: service.UniqueName(),
				Project:      service.project.Name,
				Status:       status,
				Port:         port,
			})
		}
	}
	sort.Slice(out, func(i int, j int) bool {
		return strings.Compare(out[i].Project, out[j].Project) < 0
	})
	return out, nil
}
