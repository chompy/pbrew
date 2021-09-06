package core

import (
	"bytes"
	"os/user"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
)

const nginxMainTemplateFile = "conf/nginx_main.conf.tmpl"

type nginxMainTemplate struct {
	User  string
	Group string
	Pid   string
}

func buildNginxMainTemplate() nginxMainTemplate {
	currentUser, err := user.Current()
	nginxUserName := "nginx"
	nginxGroupName := "nobody"
	if err == nil {
		nginxUserName = currentUser.Username
		grp, err := user.LookupGroupId(currentUser.Gid)
		if err == nil {
			nginxGroupName = grp.Name
		}
	}
	return nginxMainTemplate{
		User:  nginxUserName,
		Group: nginxGroupName,
		Pid:   "/tmp/pbrew-nginx.pid",
	}
}

// GenerateNginxMain returns main nginx configuration.
func GenerateNginxMain() (string, error) {
	appPath, err := appPath()
	if err != nil {
		return "", errors.WithStack(err)
	}
	templatePath := filepath.Join(appPath, nginxMainTemplateFile)
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
