package core

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"

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
	Host     string
	Original string
	Path     string
	Type     string
	Upstream string
	To       string
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
			upstream := ""
			if route.Type == "upstream" {
				service := p.MatchRelationshipToService(route.Upstream)
				switch service := service.(type) {
				case *def.App:
					{
						upstream = fmt.Sprintf("%s_%s", p.Name, service.Name)
						break
					}
				case *def.Service:
					{
						upstream = fmt.Sprintf("%s_%s", p.Name, service.Name)
						break
					}
				}
				if upstream == "" {
					continue
				}
			}
			locationTemplates = append(locationTemplates, nginxRouteLocationTemplate{
				Host:     hostName,
				Original: parsedOriginalURL.Host,
				Path:     strings.TrimRight(parsedRouteURL.Path, "/"),
				Type:     route.Type,
				Upstream: upstream,
				To:       route.To,
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
