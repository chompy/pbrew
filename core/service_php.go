package core

import (
	"strings"

	"github.com/pkg/errors"
)

// PhpExtensions is a map of PHP extension name to the command needed to install it.
type PhpExtensions map[string]string

// Match finds a PHP extension that matches given name.
func (p PhpExtensions) Match(name string) (string, string, error) {
	for extName, extCmd := range p {
		if extName == name {
			return extName, extCmd, nil
		}
	}
	for extName, extCmd := range p {
		if wildcardCompare(extName, name) {
			return extName, extCmd, nil
		}
	}
	return "", "", errors.WithStack(errors.WithMessage(ErrPHPExtNotFound, name))
}

// IsPHP returns true if service is php.
func (s *Service) IsPHP() bool {
	return strings.HasPrefix(s.BrewName, "php")
}
