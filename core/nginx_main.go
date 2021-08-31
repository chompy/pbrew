package core

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
)

const nginxMainTemplateFile = "conf/nginx_main.conf.tmpl"

type nginxMainTemplate struct {
	Pid string
}

func buildNginxMainTemplate() nginxMainTemplate {
	return nginxMainTemplate{
		Pid: "/tmp/pbrew-nginx.pid",
	}
}

// GenerateNginxMain returns main nginx configuration.
func GenerateNginxMain() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", errors.WithStack(err)
	}
	templatePath := filepath.Join(filepath.Dir(execPath), nginxMainTemplateFile)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", errors.WithStack(err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, buildNginxMainTemplate()); err != nil {
		return "", errors.WithStack(err)
	}
	return buf.String(), nil
}
