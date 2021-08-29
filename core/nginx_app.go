package core

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
)

var nginxAppTemplateFiles = map[string]string{
	"php": "conf/nginx_app_php.conf.tmpl",
}

type nginxAppTemplate struct {
	Locations []nginxAppLocationTemplate
}

type nginxAppLocationTemplate struct {
	Path     string
	Root     string
	Passthru string
	Port     int
	Env      map[string]string
}

func (p *Project) buildNginxAppTemplate(app *def.App) (nginxAppTemplate, error) {
	// get brew service info
	serviceList, err := LoadServiceList()
	if err != nil {
		return nginxAppTemplate{}, errors.WithStack(err)
	}
	service, err := serviceList.MatchDef(app)
	if err != nil {
		return nginxAppTemplate{}, errors.WithStack(err)
	}
	// build location list
	locations := make([]nginxAppLocationTemplate, 0)
	for path, location := range app.Web.Locations {
		path = strings.TrimRight(path, "/")
		root, err := filepath.Abs(filepath.Join(p.Path, location.Root))
		if err != nil {
			return nginxAppTemplate{}, errors.WithStack(err)
		}
		locations = append(locations, nginxAppLocationTemplate{
			Path:     path,
			Root:     root,
			Passthru: location.Passthru.GetString(),
			Port:     service.Port,
		})
	}
	return nginxAppTemplate{
		Locations: locations,
	}, nil
}

// GenerateNginxApp generates nginx config for given application.
func (p *Project) GenerateNginxApp(app *def.App) (string, error) {
	templatePath := nginxAppTemplateFiles[app.GetTypeName()]
	if templatePath == "" {
		return "", errors.WithStack(errors.WithMessage(ErrNginxTemplateNotFound, app.GetTypeName()))
	}
	execPath, err := os.Executable()
	if err != nil {
		return "", errors.WithStack(err)
	}
	tmpl, err := template.ParseFiles(filepath.Join(filepath.Dir(execPath), templatePath))
	if err != nil {
		return "", errors.WithStack(err)
	}
	var buf bytes.Buffer
	templateVars, err := p.buildNginxAppTemplate(app)
	if err != nil {
		return "", errors.WithStack(err)
	}
	if err := tmpl.Execute(&buf, templateVars); err != nil {
		return "", errors.WithStack(err)
	}
	return buf.String(), nil
}
