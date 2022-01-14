package core

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

// IsSolr returns true if service is solr.
func (s *Service) IsSolr() bool {
	return strings.HasPrefix(s.BrewAppName(), "java") && strings.Contains(s.StartCmd, "solr")
}

// IsSolrRunning returns true if solr is running.
func (s *Service) IsSolrRunning() bool {
	port, _ := s.Port()
	out, err := s.solrCommand("status")
	if err != nil {
		output.Warn(err.Error())
		return false
	}
	return strings.Contains(string(out), fmt.Sprintf("port %d", port))
}

func (s *Service) SolrAddConfigSets(d *def.Service, p *Project) error {
	if !s.IsSolr() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotSolr, s.BrewName))
	}
	/*for name, conf := range d.Configuration["configsets"].(map[string]interface{}) {
		output.Info(fmt.Sprintf("Create configset %s.", name))

		name = s.solrCoreName(p, name)
		path := filepath.Join(s.DataPath(), "solr", name, "configsets")
		os.MkdirAll(path, 0755)

		buf := bytes.NewBufferString(conf.(string))
		cmd := NewShellCommand()
		cmd.Env = brewEnv()
		cmd.Stdin = buf
		cmd.Args = []string{
			"-c",
			fmt.Sprintf(
				"rm -rf /tmp/solrconf && mkdir -p /tmp/solrconf && cd /tmp/solrconf && base64 -d | tar xfz - && cp -r /tmp/solrconf/* %s/",

			),
		}
		cmd.Interactive()
	}*/

	return nil
}

// SolrAddCores adds all solr cores for given project and service.
func (s *Service) SolrAddCores(d *def.Service, p *Project) error {
	if !s.IsSolr() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotSolr, s.BrewName))
	}
	for core, conf := range d.Configuration["cores"].(map[string]interface{}) {
		output.Info(fmt.Sprintf("Create core %s.", core))
		args := make([]string, 0)
		args = append(args, "-c", s.solrCoreName(p, core))
		if conf.(map[string]interface{})["conf_dir"] != nil {
			buf := bytes.NewBufferString(conf.(map[string]interface{})["conf_dir"].(string))
			cmd := NewShellCommand()
			cmd.Env = brewEnv()
			cmd.Stdin = buf
			cmd.Args = []string{
				"-c", "rm -rf /tmp/solrconf && mkdir -p /tmp/solrconf && cd /tmp/solrconf && base64 -d | tar xfz -",
			}
			cmd.Interactive()
			args = append(args, "-d", "/tmp/solrconf")
		}
		/*if conf.(map[string]interface{})["core_properties"] != nil {
			corePropPath := filepath.Join()
		}*/
		if _, err := s.solrCommand("create_core", args...); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) solrCommand(cmdStr string, args ...string) ([]byte, error) {
	var buf bytes.Buffer
	cmd := NewShellCommand()
	cmd.Env = brewEnv()
	cmd.Command = filepath.Join(GetDir(BrewDir), "opt", "solr", "bin", "solr")
	port, _ := s.Port()
	cmd.Args = make([]string, 0)
	cmd.Args = append(cmd.Args, cmdStr)
	cmd.Args = append(cmd.Args, "-p", fmt.Sprintf("%d", port))
	cmd.Args = append(cmd.Args, args...)
	cmd.Stdout = &buf
	err := cmd.Interactive()
	return buf.Bytes(), err
}

func (s *Service) solrCoreName(p *Project, core string) string {
	if p == nil {
		return core
	}
	return fmt.Sprintf("%s_%s", p.Name, core)
}

// solrPostSetup configures solr for given service definition.
func (s *Service) solrPostSetup(d *def.Service, p *Project) error {
	if !s.IsSolr() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotSolr, s.BrewName))
	}
	if err := s.SolrAddCores(d, p); err != nil {
		return err
	}
	return nil
}
