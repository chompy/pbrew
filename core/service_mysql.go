package core

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
)

const mysqlUser = "pbrew"
const mysqlPass = "pbrew"

// IsMySQL returns true if service is mysql compatible.
func (s *Service) IsMySQL() bool {
	return strings.HasPrefix(s.BrewName, "mysql") || strings.HasPrefix(s.BrewName, "mariadb")
}

// MySQLGetSchemas returns list of database schemas.
func (s *Service) MySQLGetSchemas(d *def.Service) []string {
	if !s.IsMySQL() || d == nil || d.Configuration["schemas"] == nil {
		return []string{}
	}
	schemas := d.Configuration["schemas"].([]interface{})
	out := make([]string, 0)
	for _, schema := range schemas {
		out = append(out, schema.(string))
	}
	return out
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

// MySQLDump dumps the given mysql database.
func (s *Service) MySQLDump(database string, out io.Writer) error {
	return nil
}

// MySQLExecute executes given query.
func (s *Service) MySQLExecute(query string) error {
	pathToMySQL := filepath.Join(BrewPath(), "opt", s.BrewName, "bin", "mysql")
	cmd := exec.Command(pathToMySQL, "-S", s.SocketPath(), "-e", query)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return errors.WithStack(errors.WithMessage(err, s.BrewName))
	}
	return nil
}

// mySQLSetup configures mysql for given service definition.
func (s *Service) mySQLSetup(d *def.Service, p *Project) error {
	if !s.IsMySQL() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotMySQL, s.BrewName))
	}
	// user
	output.Info(fmt.Sprintf("Create %s user.", mysqlUser))
	if err := s.MySQLExecute(fmt.Sprintf(
		"CREATE USER IF NOT EXISTS '%s'@'localhost' IDENTIFIED BY '%s';",
		mysqlUser,
		mysqlPass,
	)); err != nil {
		return errors.WithStack(errors.WithMessage(ErrServiceNotMySQL, s.BrewName))
	}
	// schemas
	schemas := s.MySQLGetSchemas(d)
	for _, schema := range schemas {
		schema = fmt.Sprintf("%s_%s", p.Name, schema)
		output.Info(fmt.Sprintf("Create %s database.", schema))
		if err := s.MySQLExecute(fmt.Sprintf(
			"CREATE SCHEMA IF NOT EXISTS %s; GRANT ALL PRIVILEGES ON %s.* TO '%s'@'localhost';",
			schema,
			schema,
			mysqlUser,
		)); err != nil {
			return errors.WithStack(errors.WithMessage(ErrServiceNotMySQL, s.BrewName))
		}
	}
	return nil
}
