package core

import (
	"encoding/base64"
	"encoding/json"

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
				"PLATFORM_RELATIONSHIPS": p.EnvPlatformRelationships(d),
				"PLATFORM_BRANCH":        "pbrew",
				"PLATFORM_ENTROPY":       "--random--",
			}
		}
	}
	return map[string]string{}
}
