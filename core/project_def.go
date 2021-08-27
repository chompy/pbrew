package core

import (
	"log"
	"strings"

	"github.com/pkg/errors"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
)

const localHostName = "localhost"

// GetDefinitionRelationship returns relationship value for given definition.
func (p *Project) GetDefinitionRelationships(d interface{}) []map[string]interface{} {
	switch d := d.(type) {
	case *def.App:
		{
			rel := d.GetEmptyRelationship()
			rel["hostname"] = localHostName
			rel["host"] = localHostName
			rel["ip"] = "127.0.0.1"
			return []map[string]interface{}{rel}
		}
	case def.Service:
		{
			brewServices, err := LoadServiceList()
			if err != nil {
				output.Warn(err.Error())
				return nil
			}
			brewService, err := brewServices.MatchDef(d)
			if err != nil {
				if !errors.Is(err, ErrServiceNotFound) {
					output.Warn(err.Error())
				}
				return nil
			}
			out := make([]map[string]interface{}, 0)
			if d.Configuration["endpoints"] != nil {
				for name, config := range d.Configuration["endpoints"].(map[string]interface{}) {
					rel := d.GetEmptyRelationship()
					rel["hostname"] = localHostName
					rel["host"] = localHostName
					rel["ip"] = "127.0.0.1"
					rel["port"] = brewService.Port
					rel["rel"] = name
					rel["path"] = config.(map[string]interface{})["default_schema"].(string)
					out = append(out, rel)
				}
			} else {
				rel := d.GetEmptyRelationship()
				rel["hostname"] = localHostName
				rel["host"] = localHostName
				rel["ip"] = "127.0.0.1"
				rel["port"] = brewService.Port
				rel["rel"] = d.GetTypeName()
				log.Println(rel)
				out = append(out, rel)
			}
			return out
		}
	}
	return nil
}

// GetRelationships gets all relationships for project.
func (p *Project) GetRelationships(d interface{}) map[string]map[string]interface{} {
	var rels map[string]string
	switch d := d.(type) {
	case *def.App:
		{
			rels = d.Relationships
			break
		}
	case *def.AppWorker:
		{
			rels = d.Relationships
			break
		}
	case def.Service:
		{
			rels = d.Relationships
			break
		}
	}
	if rels == nil {
		return nil
	}
	out := make(map[string]map[string]interface{})
	for relName, rel := range rels {
		relSplit := strings.Split(rel, ":")
		for _, service := range p.Services {
			serviceRels := p.GetDefinitionRelationships(service)
			for _, serviceRel := range serviceRels {
				if service.Name == relSplit[0] && serviceRel["rel"] == relSplit[1] {
					out[relName] = serviceRel
					break
				}
			}
		}
	}
	return out
}

// GetDefinitionRelationships returns relationships for given definition.
/*func (p *Project) GetDefinitionRelationships(d interface{}) map[string][]map[string]interface{} {
	var rels map[string]string
	switch d := d.(type) {
	case *def.App:
		{
			rels = d.Relationships
			break
		}
	case *def.AppWorker:
		{
			rels = d.Relationships
			break
		}
	case def.Service:
		{
			rels = d.Relationships
			break
		}
	}
	out := make(map[string][]map[string]interface{})
	for name, rel := range rels {
		out[name] = make([]map[string]interface{}, 0)
		relSplit := strings.Split(rel, ":")

		for _, service := range p.Services {
			if service.Name == relSplit[0] && (
		}

		for _, v := range p.relationships {
			if v["service"] != nil && v["service"].(string) == relSplit[0] && v["rel"] == relSplit[1] {
				out[name] = append(out[name], v)
			}
		}
	}
	return out
}
*/
