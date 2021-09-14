package core

import (
	"bytes"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
)

const phpFpmPoolTemplateFile = "conf/php_fpm_pool.conf.tmpl"

type phpFpmPoolTemplate struct {
	ProjectName string
	AppName     string
	User        string
	Socket      string
	Env         map[string]string
	Ini         map[string]string
}

func (p *Project) buildPhpFPMPoolTemplate(app *def.App) (phpFpmPoolTemplate, error) {
	serviceList, err := LoadServiceList()
	if err != nil {
		return phpFpmPoolTemplate{}, err
	}
	service, err := serviceList.MatchDef(app)
	if err != nil {
		return phpFpmPoolTemplate{}, err
	}
	vars, err := p.Variables(app)
	if err != nil {
		return phpFpmPoolTemplate{}, err
	}
	return phpFpmPoolTemplate{
		ProjectName: p.Name,
		AppName:     app.Name,
		Socket:      service.UpstreamSocketPath(p, app),
		Env:         p.Env(app),
		Ini:         vars.GetStringSubMap("php"),
	}, nil
}

// GenerateNginxApp generates nginx config for given application.
func (p *Project) GeneratePhpFpmPool(app *def.App) (string, error) {
	tmpl, err := template.ParseFiles(filepath.Join(GetDir(AppDir), phpFpmPoolTemplateFile))
	if err != nil {
		return "", errors.WithStack(err)
	}
	var buf bytes.Buffer
	templateVars, err := p.buildPhpFPMPoolTemplate(app)
	if err != nil {
		return "", errors.WithStack(err)
	}
	if err := tmpl.Execute(&buf, templateVars); err != nil {
		return "", errors.WithStack(err)
	}
	return buf.String(), nil
}
