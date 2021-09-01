package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"github.com/pkg/errors"
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

// MySQLShell enters the mysql shell.
func (s *Service) MySQLShell(database string) error {
	if !s.IsMySQL() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotMySQL, s.BrewName))
	}
	if !s.IsRunning() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotRunning, s.BrewName))
	}
	output.Info(fmt.Sprintf("Access shell for %s.", s.BrewName))
	pathToMySQL := filepath.Join(BrewPath(), "opt", s.BrewName, "bin", "mysql")
	args := make([]string, 0)
	args = append(args, "-S", s.SocketPath())
	if database != "" {
		args = append(args, "-D", database)
	}
	cmd := exec.Command(pathToMySQL, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return errors.WithStack(errors.WithMessage(err, s.BrewName))
	}
	return nil
}
