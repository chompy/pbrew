package core

import (
	"sort"
	"strings"
)

const serviceStatusNotInstalled = "not installed"
const serviceStatusStopped = "stopped"
const serviceStatusRunning = "running"

// ServiceStatus defines
type ServiceStatus struct {
	Name     string   `json:"name"`
	Ports    []int    `json:"ports"`
	Projects []string `json:"projects"`
	Status   string   `json:"status"`
}

// GetServiceStatuses returns status of all services.
func GetServiceStatuses() ([]ServiceStatus, error) {
	brewServices, err := LoadServiceList()
	if err != nil {
		return nil, err
	}
	portMaps, err := LoadPortMap()
	if err != nil {
		return nil, err
	}
	projectTracks, err := ProjectTrackGet()
	if err != nil {
		return nil, err
	}
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	nginxService := NginxService()
	brewServices[nginxService.BrewAppName()] = nginxService
	// list project statuses
	out := make([]ServiceStatus, 0)
	for _, service := range brewServices {
		// check if already listed
		alreadyCreated := false
		for _, addedService := range out {
			if addedService.Name == service.BrewAppName() {
				alreadyCreated = true
				break
			}
		}
		if alreadyCreated {
			continue
		}
		// default status
		status := serviceStatusNotInstalled
		// get projects + ports + status
		projects := make([]string, 0)
		ports := make([]int, 0)
		for _, pt := range projectTracks {
			for _, ptService := range pt.Services {
				if ptService == service.BrewAppName() {
					projects = append(projects, pt.Name)
					service.ProjectName = pt.Name
					if service.IsRunning() {
						status = serviceStatusRunning
					}
					port, err := portMaps.ServicePort(service)
					if err != nil {
						return nil, err
					}
					hasPort := false
					for _, existingPort := range ports {
						if existingPort == port {
							hasPort = true
							break
						}
					}
					if !hasPort {
						ports = append(ports, port)
					}
					break
				}
			}
		}

		if service.IsRunning() {
			status = serviceStatusRunning
		} else if service.IsInstalled() && status == serviceStatusNotInstalled {
			status = serviceStatusStopped
		}

		if service.BrewAppName() == nginxService.BrewAppName() {
			ports = []int{config.RouterHTTP, config.RouterHTTPS}
		}
		if service.BrewAppName() == nginxService.BrewAppName() {
			for _, pt := range projectTracks {
				proj := &Project{Name: pt.Name}
				if NginxHas(proj) {
					projects = append(projects, pt.Name)
				}
			}
		}
		out = append(out, ServiceStatus{
			Name:     service.BrewAppName(),
			Ports:    ports,
			Status:   status,
			Projects: projects,
		})
	}
	sort.Slice(out, func(i int, j int) bool {
		return strings.Compare(out[i].Name, out[j].Name) < 0
	})
	return out, nil
}
