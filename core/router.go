package core

import (
	"net/url"
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

const defaultHostName = "localhost"

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

/*func GetUpstreamHost(proj *project.Project, upstream string, allowServices bool) (string, error) {
	upstreamSplit := strings.Split(upstream, ":")
	// itterate apps and services to find name match
	// TODO this should use relationships but those only get resolved when
	// services are opened...sooo??
	for _, app := range proj.Apps {
		if app.Name == upstreamSplit[0] {
			return proj.GetDefinitionHostName(app), nil
		}
	}
	for _, serv := range proj.Services {
		if serv.Name == upstreamSplit[0] {
			// forward to app if allowServices is false
			if !allowServices {
				for _, relationship := range serv.Relationships {
					rlSplit := strings.Split(relationship, ":")
					return GetUpstreamHost(proj, fmt.Sprintf("%s:http", rlSplit[0]), allowServices)
				}
			}
			// TODO use relationship to determine port
			port := 80
			switch serv.GetTypeName() {
			case "varnish", "solr":
				{
					port = 8080
					break
				}
			}
			return fmt.Sprintf("%s:%d", proj.GetDefinitionHostName(serv), port), nil
		}
	}
	return "", errors.Wrapf(ErrUpstreamNotFound, "upstream %s not found", upstream)
}
*/
