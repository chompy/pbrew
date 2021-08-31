package core

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

const defaultHostName = "localhost"

// NginxService returns the service for nginx.
func NginxService() *Service {
	nginxConf, err := GenerateNginxMain()
	if err != nil {
		output.Warn(err.Error())
		return nil
	}
	nginxConfB64 := base64.StdEncoding.EncodeToString([]byte(nginxConf))
	return &Service{
		BrewName:       "nginx",
		PostInstallCmd: "",
		StartCmd: fmt.Sprintf(
			"echo '%s' > {BREW_PATH}/etc/nginx/nginx.conf && {BREW_PATH}/opt/nginx/bin/nginx",
			nginxConfB64,
		),
		StopCmd: "{BREW_PATH}/opt/nginx/bin/nginx -s stop",
		Port:    8080,
	}
}

// NginxAdd generates nginx config for given project.
func NginxAdd(proj *Project) error {
	if proj == nil {
		return nil
	}
	done := output.Duration(fmt.Sprintf("Add '%s' to router.", proj.Name))
	nginxRoutes, err := proj.GenerateNginxRoutes()
	if err != nil {
		return errors.WithStack(err)
	}
	if err := ioutil.WriteFile(
		filepath.Join(BrewPath(), "etc", "nginx", "servers", proj.Name),
		[]byte(nginxRoutes),
		0655,
	); err != nil {
		return errors.WithStack(err)
	}
	os.MkdirAll(filepath.Join(BrewPath(), "etc", "nginx", "upstreams"), 0755)
	for _, app := range proj.Apps {
		nginxApp, err := proj.GenerateNginxApp(app)
		if err != nil {
			return errors.WithStack(err)
		}
		if err := ioutil.WriteFile(
			filepath.Join(BrewPath(), "etc", "nginx", "upstreams", proj.Name+"_"+app.Name),
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
