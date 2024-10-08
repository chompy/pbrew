package core

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
)

const mysqlUser = "pbrew"
const mysqlPass = "pbrew"

// IsMySQL returns true if service is mysql compatible.
func (s *Service) IsMySQL() bool {
	return strings.HasPrefix(s.BrewAppName(), "mysql") || strings.HasPrefix(s.BrewAppName(), "mariadb")
}

// MySQLGetSchemas returns list of database schemas.
func (s *Service) MySQLGetSchemas() []string {
	switch d := s.definition.(type) {
	case *def.Service:
		{
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
	}
	return []string{}
}

// MySQLShell enters the mysql shell.
func (s *Service) MySQLShell(database string) error {
	if !s.IsMySQL() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotMySQL, s.DisplayName()))
	}
	if !s.IsRunning() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotRunning, s.DisplayName()))
	}
	output.Info(fmt.Sprintf("Access shell for %s.", s.DisplayName()))
	pathToMySQL := filepath.Join(GetDir(BrewDir), "opt", s.BrewAppName(), "bin", "mysql")
	args := make([]string, 0)
	args = append(args, "-S", s.SocketPath(), "-u", "root")
	if database != "" {
		args = append(args, "-D", database)
	}
	cmd := NewShellCommand()
	cmd.Command = pathToMySQL
	cmd.Args = args
	if err := cmd.Drop(); err != nil {
		return errors.WithStack(errors.WithMessage(err, s.DisplayName()))
	}
	return nil
}

// MySQLDump dumps the given mysql database.
func (s *Service) MySQLDump(database string) error {
	pathToMySQL := filepath.Join(GetDir(BrewDir), "opt", s.BrewAppName(), "bin", "mysqldump")
	cmd := NewShellCommand()
	cmd.Command = pathToMySQL
	cmd.Args = []string{"-S", s.SocketPath(), "-u", "root", database}
	if err := cmd.Drop(); err != nil {
		return errors.WithStack(errors.WithMessage(err, s.DisplayName()))
	}
	return nil
}

// MySQLExecute executes given query.
func (s *Service) MySQLExecute(query string) error {
	pathToMySQL := filepath.Join(GetDir(BrewDir), "opt", s.BrewAppName(), "bin", "mysql")
	cmd := NewShellCommand()
	cmd.Command = pathToMySQL
	cmd.Args = []string{"-S", s.SocketPath(), "-u", "root", "-e", query}
	if err := cmd.Interactive(); err != nil {
		return errors.WithStack(errors.WithMessage(err, s.DisplayName()))
	}
	return nil
}

func (s *Service) mySQLSchemeName(name string) string {
	if s.project == nil {
		return name
	}
	return fmt.Sprintf("%s_%s", strings.ReplaceAll(s.project.Name, "-", "_"), name)
}

// mySQLPostSetup configures mysql for given service definition.
func (s *Service) mySQLPostSetup() error {
	if !s.IsMySQL() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotMySQL, s.DisplayName()))
	}
	// user
	// TODO project specific user+password
	output.Info(fmt.Sprintf("Create %s user.", mysqlUser))
	if err := s.MySQLExecute(fmt.Sprintf(
		"CREATE USER IF NOT EXISTS '%s'@'localhost' IDENTIFIED BY '%s';",
		mysqlUser,
		mysqlPass,
	)); err != nil {
		return errors.WithStack(errors.WithMessage(ErrServiceNotMySQL, s.DisplayName()))
	}
	// schemas
	schemas := s.MySQLGetSchemas()
	for _, schema := range schemas {
		schema = s.mySQLSchemeName(schema)
		output.Info(fmt.Sprintf("Create %s database.", schema))
		if err := s.MySQLExecute(fmt.Sprintf(
			"CREATE SCHEMA IF NOT EXISTS %s; GRANT ALL PRIVILEGES ON %s.* TO '%s'@'localhost';",
			schema,
			schema,
			mysqlUser,
		)); err != nil {
			return errors.WithStack(errors.WithMessage(ErrServiceNotMySQL, s.DisplayName()))
		}
	}
	return nil
}

func (s *Service) mySQLPurge() error {
	if !s.IsMySQL() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotMySQL, s.DisplayName()))
	}
	// needs to be running to drop schemas
	wasRunning := s.IsRunning()
	if !wasRunning {
		if err := s.Start(); err != nil {
			return err
		}
		time.Sleep(time.Second * 3)
	}
	// schemas
	schemas := s.MySQLGetSchemas()
	for _, schema := range schemas {
		schema = s.mySQLSchemeName(schema)
		output.Info(fmt.Sprintf("Drop %s database.", schema))
		if err := s.MySQLExecute(fmt.Sprintf(
			"DROP SCHEMA IF EXISTS %s;",
			schema,
		)); err != nil {
			return errors.WithStack(errors.WithMessage(ErrServiceNotMySQL, s.DisplayName()))
		}
	}
	// stop if it wasn't running
	if !wasRunning {
		if err := s.Stop(); err != nil {
			return err
		}
	}
	return nil
}
