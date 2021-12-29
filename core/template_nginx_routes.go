package core

import (
	"bytes"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"text/template"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"github.com/pkg/errors"
)

const nginxRouteTemplateFile = "conf/nginx_routes.conf.tmpl"

type nginxRouteTemplate struct {
	ProjectName string
	Hosts       []nginxRouteHostTemplate
}

type nginxRouteHostTemplate struct {
	Host      string
	PortHTTP  int
	PortHTTPS int
	Locations []nginxRouteLocationTemplate
	ErrorLog  string
	AccessLog string
	DataDir   string
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
	config, err := LoadConfig()
	if err != nil {
		output.Warn(err.Error())
		return nginxRouteTemplate{}
	}
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
			path := strings.TrimRight(parsedRouteURL.Path, "/")
			hasPath := false
			for _, l := range locationTemplates {
				if l.Path == path {
					hasPath = true
					break
				}
			}
			if hasPath {
				continue
			}
			locationTemplates = append(locationTemplates, nginxRouteLocationTemplate{
				Host:         ProjectDefaultHostName(p, hostName),
				Original:     parsedOriginalURL.Host,
				Path:         path,
				Type:         route.Type,
				UpstreamPort: upstreamPort,
				To:           route.To,
			})
		}

		hostTemplates = append(hostTemplates, nginxRouteHostTemplate{
			Host:      ProjectDefaultHostName(p, hostName),
			PortHTTP:  config.RouterHTTP,
			PortHTTPS: config.RouterHTTPS,
			Locations: locationTemplates,
			ErrorLog:  filepath.Join(GetDir(LogDir), fmt.Sprintf("nginx_error_%s.log", p.Name)),
			AccessLog: filepath.Join(GetDir(LogDir), fmt.Sprintf("nginx_access_%s.log", p.Name)),
			DataDir:   filepath.Join(GetDir(DataDir), "nginx"),
		})
	}
	return nginxRouteTemplate{
		ProjectName: p.Name,
		Hosts:       hostTemplates,
	}
}

// GenerateNginxRoutes returns nginx configuration for project routes.
func (p *Project) GenerateNginxRoutes() (string, error) {
	templatePath := filepath.Join(GetDir(AppDir), nginxRouteTemplateFile)
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
