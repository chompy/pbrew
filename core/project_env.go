package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

// EnvPlatformRelationships returns PLATFORM_RELATIONSHIPS environment variable.
func (p *Project) EnvPlatformRelationships(d interface{}) string {
	rels := p.MapRelationships(d)
	jsonRaw, err := json.Marshal(rels)
	if err != nil {
		output.Warn(err.Error())
		return ""
	}
	return base64.StdEncoding.EncodeToString(jsonRaw)
}

// EnvPlatformVariables returns PLATFORM_VARIABLES environment variable.
func (p *Project) EnvPlatformVariables(d interface{}) string {
	vars, err := p.Variables(d)
	if err != nil {
		output.Warn(err.Error())
		return ""
	}
	for _, k := range vars.Keys() {
		if strings.HasPrefix(k, "env") {
			vars.Delete(k)
		}
	}
	jsonOut, err := json.Marshal(vars)
	if err != nil {
		output.Warn(err.Error())
		return ""
	}
	return base64.StdEncoding.EncodeToString(jsonOut)
}

// EnvPlatformRoutes returns PLATFORM_ROUTES environment variable.
func (p *Project) EnvPlatformRoutes(d interface{}) string {
	replaceDefault := func(path string) string {
		return strings.ReplaceAll(path, "{default}", fmt.Sprintf("%s.default", p.Name))
	}
	routes := make(map[string]def.Route)
	for _, route := range p.Routes {
		route.OriginalURL = replaceDefault(route.OriginalURL)
		route.Path = replaceDefault(route.Path)
		route.To = replaceDefault(route.To)
		for i := range route.Redirects.Paths {
			route.Redirects.Paths[i].To = replaceDefault(route.Redirects.Paths[i].To)
		}
		for k, v := range route.Attributes {
			route.Attributes[k] = replaceDefault(v)
		}
		routes[route.Path] = route
	}
	jsonOut, err := json.Marshal(routes)
	if err != nil {
		output.Warn(err.Error())
		return ""
	}
	return base64.StdEncoding.EncodeToString(jsonOut)
}

// Env returns all environment variables for given service definition.
func (p *Project) Env(d interface{}) map[string]string {
	switch d := d.(type) {
	case *def.App:
		{
			vars, err := p.Variables(d)
			if err != nil {
				output.Warn(err.Error())
				return nil
			}
			out := vars.GetStringSubMap("env")
			for k, v := range out {
				if strings.Contains(v, "\n") {
					out[k] = ""
				}
			}
			out["PLATFORM_APP_DIR"] = p.Path
			out["PLATFORM_DIR"] = p.Path
			out["PLATFORM_DOCUMENT_ROOT"] = ""
			out["PLATFORM_BRANCH"] = "pcc"
			out["PLATFORM_PROJECT"] = p.Name
			out["PLATFORM_PROJECT_ENTROPY"] = "--random--"
			out["PLATFORM_ENVIRONMENT"] = fmt.Sprintf("pbrew-%s", p.Name)
			out["PLATFORM_ENVIRONMENT_TYPE"] = "dev"
			out["PLATFORM_APPLICATION_NAME"] = d.Name
			out["PLATFORM_APP_COMMAND"] = ""
			out["PLATFORM_RELATIONSHIPS"] = p.EnvPlatformRelationships(d)
			out["PLATFORM_ROUTES"] = p.EnvPlatformRoutes(d)
			out["PLATFORM_VARIABLES"] = p.EnvPlatformVariables(d)
			return out
		}
	}
	return map[string]string{}
}
