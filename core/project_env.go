package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

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
	return base64.RawStdEncoding.EncodeToString(jsonRaw)
}

// Env returns all environment variables for given service definition.
func (p *Project) Env(d interface{}) map[string]string {
	switch d := d.(type) {
	case *def.App:
		{
			return map[string]string{
				"PLATFORM_APP_DIR":          p.Path,
				"PLATFORM_DIR":              p.Path,
				"PLATFORM_DOCUMENT_ROOT":    "",
				"PLATFORM_BRANCH":           "pbrew",
				"PLATFORM_PROJECT":          p.Name,
				"PLATFORM_PROJECT_ENTROPY":  "--random--",
				"PLATFORM_ENVIRONMENT":      fmt.Sprintf("pbrew-%s", p.Name),
				"PLATFORM_ENVIRONMENT_TYPE": "dev",
				"PLATFORM_APPLICATION_NAME": d.Name,
				"PLATFORM_APP_COMMAND":      "",
				"PLATFORM_RELATIONSHIPS":    p.EnvPlatformRelationships(d),
				"PLATFORM_ROUTES":           "",
				"PLATFORM_VARIABLES":        "",
			}
		}
	}
	return map[string]string{}
}
