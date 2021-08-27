package core

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"text/template"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"

	"github.com/pkg/errors"
)

const nginxRouteTemplate = "conf/nginx_routes.conf.tmpl"

type routeTemplate struct {
	ProjectName string
	Hosts       []routeHostTemplate
}

type routeHostTemplate struct {
	Host      string
	Port      int
	Locations []routeLocationTemplate
}

type routeLocationTemplate struct {
	Host     string
	Path     string
	Type     string
	Upstream string
	To       string
}

func (p *Project) buildRouteTemplate() routeTemplate {
	hostTemplates := make([]routeHostTemplate, 0)
	for _, hostName := range GetHostNames(p.Routes) {
		locationTemplates := make([]routeLocationTemplate, 0)
		for _, route := range GetRoutesForHostName(hostName, p.Routes) {
			parsedRouteURL, err := url.Parse(route.Path)
			if err != nil {
				continue
			}
			upstream := ""
			if route.Type == "upstream" {
				service := p.MatchRelationshipToService(route.Upstream)
				log.Println(route.Upstream)
				switch service := service.(type) {
				case *def.App:
					{
						upstream = fmt.Sprintf("%s_%s.conf", p.Name, service.Name)
						break
					}
				case *def.Service:
					{
						upstream = fmt.Sprintf("%s_%s.conf", p.Name, service.Name)
						break
					}
				}
			}
			locationTemplates = append(locationTemplates, routeLocationTemplate{
				Host:     hostName,
				Path:     parsedRouteURL.Path,
				Type:     route.Type,
				Upstream: upstream,
			})
		}
		hostTemplates = append(hostTemplates, routeHostTemplate{
			Host:      hostName,
			Port:      8080,
			Locations: locationTemplates,
		})
	}
	return routeTemplate{
		ProjectName: p.Name,
		Hosts:       hostTemplates,
	}
}

// GenerateNginxRoutes returns nginx configuration for project routes.
func (p *Project) GenerateNginxRoutes() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", errors.WithStack(err)
	}
	templatePath := filepath.Join(filepath.Dir(execPath), nginxRouteTemplate)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", errors.WithStack(err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, p.buildRouteTemplate()); err != nil {
		return "", errors.WithStack(err)
	}
	return buf.String(), nil
}
