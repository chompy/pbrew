package core

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

const nginxRouteTemplateFile = "conf/nginx_routes.conf.tmpl"

type nginxRouteTemplate struct {
	ProjectName string
	Hosts       []nginxRouteHostTemplate
}

type nginxRouteHostTemplate struct {
	Host      string
	Port      int
	Locations []nginxRouteLocationTemplate
}

type nginxRouteLocationTemplate struct {
	Host         string
	Original     string
	Path         string
	Type         string
	UpstreamPort int
	To           string
}

func (p *Project) buildNginxRouteTemplate() nginxRouteTemplate {
	hostTemplates := make([]nginxRouteHostTemplate, 0)
	for _, hostName := range GetHostNames(p.Routes) {
		locationTemplates := make([]nginxRouteLocationTemplate, 0)
		for _, route := range GetRoutesForHostName(hostName, p.Routes) {
			if route.Path != "" && route.Path[0] == '.' {
				continue
			}
			parsedRouteURL, err := url.Parse(route.Path)
			if err != nil {
				continue
			}
			parsedOriginalURL, err := url.Parse(route.OriginalURL)
			if err != nil {
				continue
			}
			upstreamPort := 0
			if route.Type == "upstream" {
				service := p.MatchRelationshipToService(route.Upstream)
				if service != nil {
					upstreamPort = p.GetUpstreamPort(service)
				}
				if upstreamPort == 0 {
					continue
				}
			}
			locationTemplates = append(locationTemplates, nginxRouteLocationTemplate{
				Host:         hostName,
				Original:     parsedOriginalURL.Host,
				Path:         strings.TrimRight(parsedRouteURL.Path, "/"),
				Type:         route.Type,
				UpstreamPort: upstreamPort,
				To:           route.To,
			})
		}
		hostTemplates = append(hostTemplates, nginxRouteHostTemplate{
			Host:      hostName,
			Port:      8080,
			Locations: locationTemplates,
		})
	}
	return nginxRouteTemplate{
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
	templatePath := filepath.Join(filepath.Dir(execPath), nginxRouteTemplateFile)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", errors.WithStack(err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, p.buildNginxRouteTemplate()); err != nil {
		return "", errors.WithStack(err)
	}
	return buf.String(), nil
}
