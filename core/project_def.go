package core

import (
	"log"
	"strings"

	"github.com/pkg/errors"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
)

const localHostName = "localhost"

// GenerateRelationships returns available relationship mappings for given service definition.
func GenerateRelationships(d interface{}) []map[string]interface{} {
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

// MapRelationships returns all relationships for given service definition.
func (p *Project) MapRelationships(d interface{}) map[string]map[string]interface{} {
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
			serviceRels := GenerateRelationships(service)
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

// MatchRelationshipToService matches a given relationship to its service def.
func (p *Project) MatchRelationshipToService(rel string) interface{} {
	relSplit := strings.Split(rel, ":")
	// look for service match
	for _, service := range p.Services {
		if service.Name == relSplit[0] {
			serviceRels := GenerateRelationships(service)
			for _, serviceRel := range serviceRels {
				if serviceRel["rel"] == relSplit[1] {
					return service
				}
			}
		}
	}
	// map varnish to first app
	if relSplit[0] == "varnish" {
		return p.Apps[0]
	}
	// look for app match
	for _, service := range p.Apps {
		if service.Name == relSplit[0] {
			serviceRels := GenerateRelationships(service)
			for _, serviceRel := range serviceRels {
				if serviceRel["rel"] == relSplit[1] {
					return service
				}
			}
		}
	}
	return nil
}
