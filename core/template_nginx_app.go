package core

import (
	"bytes"
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
	Port      int
	Locations []nginxAppLocationTemplate
}

type nginxAppLocationTemplate struct {
	Path     string
	Root     string
	Passthru string
	Socket   string
}

func (p *Project) buildNginxAppTemplate(app *def.App) (nginxAppTemplate, error) {
	// get brew service info
	serviceList, err := LoadServiceList()
	if err != nil {
		return nginxAppTemplate{}, err
	}
	service, err := serviceList.MatchDef(app)
	if err != nil {
		return nginxAppTemplate{}, err
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
			Socket:   service.UpstreamSocketPath(p, app),
		})
	}
	return nginxAppTemplate{
		Port:      p.GetUpstreamPort(app),
		Locations: locations,
	}, nil
}

// GenerateNginxApp generates nginx config for given application.
func (p *Project) GenerateNginxApp(app *def.App) (string, error) {
	templatePath := nginxAppTemplateFiles[app.GetTypeName()]
	if templatePath == "" {
		return "", errors.WithStack(errors.WithMessage(ErrTemplateNotFound, app.GetTypeName()))
	}

	tmpl, err := template.ParseFiles(filepath.Join(GetDir(AppDir), templatePath))
	if err != nil {
		return "", errors.WithStack(err)
	}
	var buf bytes.Buffer
	templateVars, err := p.buildNginxAppTemplate(app)
	if err != nil {
		return "", err
	}
	if err := tmpl.Execute(&buf, templateVars); err != nil {
		return "", errors.WithStack(err)
	}
	return buf.String(), nil
}
