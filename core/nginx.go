package core

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

const defaultHostName = "localhost"

const nginxStartCmd = `
	mkdir -p /tmp/pbrew_nginx
	cp {APP_PATH}/conf/nginx_fastcgi_params.normal {CONF_PATH}/nginx_fastcgi_params.normal
	sudo {BREW_PATH}/opt/nginx/bin/nginx -c {CONF_FILE} -p {BREW_PATH}/opt/nginx/ -e {LOG_PATH}/nginx_error.log
`

const nginxPostInstallCmd = `
	openssl req -x509 -out {DATA_PATH}/localhost.crt -keyout {DATA_PATH}/localhost.key \
		-newkey rsa:2048 -nodes -sha256 \
		-subj '/CN=localhost' -extensions EXT -config <( \
		printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")
`

// NginxService returns the service for nginx.
func NginxService() *Service {
	return &Service{
		BrewName:        "nginx",
		PostInstallCmd:  nginxPostInstallCmd,
		StartCmd:        nginxStartCmd,
		StopCmd:         "sudo {BREW_PATH}/opt/nginx/bin/nginx -c {CONF_FILE} -p {BREW_PATH}/opt/nginx/ -e {LOG_PATH}/nginx_error.log -s stop",
		ReloadCmd:       "sudo {BREW_PATH}/opt/nginx/bin/nginx -c {CONF_FILE} -p {BREW_PATH}/opt/nginx/ -e {LOG_PATH}/nginx_error.log -s reload",
		ConfigTemplates: map[string]string{"nginx_main.conf.tmpl": "{CONF_FILE}"},
	}
}

// NginxRouteConfigPath returns path for project route config.
func NginxRouteConfigPath(p *Project) string {
	return filepath.Join(GetDir(ConfDir), fmt.Sprintf("nginx_routes_%s.conf", p.Name))
}

// NginxAppConfigPath returns path for application config.
func NginxAppConfigPath(p *Project, def *def.App) string {
	return filepath.Join(GetDir(ConfDir), fmt.Sprintf("nginx_app_%s_%s.conf", p.Name, def.Name))
}

// NginxAdd generates nginx config for given project.
func NginxAdd(proj *Project) error {
	if proj == nil {
		return nil
	}
	done := output.Duration(fmt.Sprintf("Add '%s' to router.", proj.Name))
	nginxRoutes, err := proj.GenerateNginxRoutes()
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(
		NginxRouteConfigPath(proj),
		[]byte(nginxRoutes),
		0655,
	); err != nil {
		return errors.WithStack(err)
	}
	for _, app := range proj.Apps {
		nginxApp, err := proj.GenerateNginxApp(app)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(
			NginxAppConfigPath(proj, app),
			[]byte(nginxApp),
			0655,
		); err != nil {
			return errors.WithStack(err)
		}
	}
	done()
	return nil
}

// NginxDel deletes nginx config for given project.
func NginxDel(proj *Project) error {
	return nil
}

// GetHostNames returns all host names for given routes.
func GetHostNames(routes []def.Route) []string {
	out := make([]string, 0)
	for _, route := range routes {
		urlParse, err := url.Parse(route.Path)
		if err != nil {
			output.LogWarn(err.Error())
			continue
		}
		thisHost := strings.TrimSpace(urlParse.Host)
		if thisHost == "" {
			thisHost = defaultHostName
		}

		hasHost := false
		for _, host := range out {
			if host == thisHost {
				hasHost = true
				break
			}
		}
		if !hasHost {
			out = append(out, thisHost)
		}
	}
	return out
}

// GetRoutesForHostName returns all routes for given host name.
func GetRoutesForHostName(host string, routes []def.Route) []def.Route {
	out := make([]def.Route, 0)
	for _, route := range routes {
		urlParse, err := url.Parse(route.Path)
		if err != nil {
			output.LogWarn(err.Error())
			continue
		}
		thisHost := strings.TrimSpace(urlParse.Host)
		if thisHost == "" {
			thisHost = defaultHostName
		}
		if thisHost != host {
			continue
		}
		out = append(out, route)
	}
	return out
}
