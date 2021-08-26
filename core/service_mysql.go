package core

import (
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
)

// IsMySQL returns true if service is mysql compatible.
func (s *Service) IsMySQL() bool {
	return strings.HasPrefix(s.BrewName, "mysql") || strings.HasPrefix(s.BrewName, "mariadb")
}

// MySQLGetSchemas returns list of database schemas.
func (s *Service) MySQLGetSchemas(d *def.Service) []string {
	if !s.IsMySQL() || d == nil || d.Configuration["schemas"] == nil {
		return []string{}
	}
	return d.Configuration["schemas"].([]string)
}
