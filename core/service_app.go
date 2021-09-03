package core

import "strings"

// IsPHP returns true if service is php.
func (s *Service) IsPHP() bool {
	return strings.HasPrefix(s.BrewName, "php")
}
