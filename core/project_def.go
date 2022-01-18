package core

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
)

// GenerateRelationships returns available relationship mappings for given service definition.
func (p *Project) GenerateRelationships(d interface{}) []map[string]interface{} {
	switch d := d.(type) {
	case *def.App:
		{
			rel := d.GetEmptyRelationship()
			rel["ip"] = "127.0.0.1"
			rel["hostname"] = rel["ip"]
			rel["host"] = rel["ip"]
			if rel["rel"] == "" {
				rel["rel"] = "http"
			}
			return []map[string]interface{}{rel}
		}
	case def.Service:
		{
			brewServices, err := LoadServiceList()
			if err != nil {
				output.Warn(err.Error())
				return nil
			}
			port := 0

			// load service
			brewService, err := brewServices.MatchDef(d)
			if err != nil {
				if !errors.Is(err, ErrServiceNotFound) {
					output.Warn(err.Error())
					return nil
				}
			}
			if brewService != nil {
				brewService.ProjectName = p.Name
				port, err = brewService.Port()
				if err != nil {
					output.Warn(err.Error())
					return nil
				}
			}

			// load service override
			serviceOverride, err := MatchServiceOverrideDef(d)
			if err != nil {
				if !errors.Is(err, ErrServiceNotFound) {
					output.Warn(err.Error())
					return nil
				}
			}

			out := make([]map[string]interface{}, 0)
			if d.Configuration["endpoints"] != nil {
				for name, config := range d.Configuration["endpoints"].(map[string]interface{}) {
					rel := d.GetEmptyRelationship()
					rel["ip"] = "127.0.0.1"
					rel["hostname"] = rel["ip"]
					rel["host"] = rel["ip"]
					rel["port"] = port
					rel["rel"] = name
					if strings.HasPrefix(name, "redis") {
						rel["rel"] = "redis"
					}
					if serviceOverride != nil {
						rel = serviceOverride.Relationship()
						rel["rel"] = name
					} else if brewService != nil && brewService.IsMySQL() {
						rel["path"] = p.ResolveDatabase(config.(map[string]interface{})["default_schema"].(string))
						rel["username"] = mysqlUser
						rel["password"] = mysqlPass
						rel["scheme"] = "mysql"
						rel["query"] = map[string]interface{}{
							"is_master": true,
						}
					} else if brewService != nil && brewService.IsSolr() {
						rel["path"] = fmt.Sprintf("solr/%s", brewService.SolrCoreName(p, name))
						rel["scheme"] = "solr"
					}
					out = append(out, rel)
				}
			} else {
				rel := d.GetEmptyRelationship()
				rel["ip"] = "127.0.0.1"
				rel["hostname"] = rel["ip"]
				rel["host"] = rel["ip"]
				rel["port"] = port
				rel["rel"] = d.GetTypeName()
				if strings.HasPrefix(rel["rel"].(string), "redis") {
					rel["rel"] = "redis"
				}
				out = append(out, rel)
			}
			return out
		}
	}
	return nil
}

// MapRelationships returns all relationships for given service definition.
func (p *Project) MapRelationships(d interface{}) map[string][]map[string]interface{} {
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
	out := make(map[string][]map[string]interface{})
	for relName, rel := range rels {
		relSplit := strings.Split(rel, ":")
		out[relName] = make([]map[string]interface{}, 0)
		for _, service := range p.Services {
			serviceRels := p.GenerateRelationships(service)
			for _, serviceRel := range serviceRels {
				if service.Name == relSplit[0] && serviceRel["rel"] == relSplit[1] {
					out[relName] = append(out[relName], serviceRel)
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
			serviceRels := p.GenerateRelationships(service)
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
			serviceRels := p.GenerateRelationships(service)
			for _, serviceRel := range serviceRels {
				if serviceRel["rel"] == relSplit[1] {
					return service
				}
			}
		}
	}
	return nil
}

// ResolveDatabase returns actual database name from endpoint.
func (p *Project) ResolveDatabase(database string) string {
	if database != "" && !strings.HasPrefix(database, p.Name) {
		database = fmt.Sprintf("%s_%s", strings.ReplaceAll(p.Name, "-", "_"), database)
	}
	return database
}
