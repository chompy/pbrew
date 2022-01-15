package core

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

// IsSolr returns true if service is solr.
func (s *Service) IsSolr() bool {
	return s.Name == "solr"
}

// IsSolrRunning returns true if solr is running.
func (s *Service) IsSolrRunning() bool {
	port, _ := s.Port()
	out, _ := s.solrCommand("status")
	return strings.Contains(string(out), fmt.Sprintf("port %d", port))
}

// SolrAddConfigSets adds configsets defines in given project.
func (s *Service) SolrAddConfigSets(d *def.Service, p *Project) error {
	if !s.IsSolr() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotSolr, s.DisplayName()))
	}
	// TODO
	/*for name, conf := range d.Configuration["configsets"].(map[string]interface{}) {
		output.Info(fmt.Sprintf("Create configset %s.", name))

		name = s.solrCoreName(p, name)
		path := filepath.Join(s.DataPath(), name, "configsets")
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
		return errors.WithStack(errors.WithMessage(ErrServiceNotSolr, s.DisplayName()))
	}
	for core, conf := range d.Configuration["cores"].(map[string]interface{}) {
		if s.SolrHasCore(p, core) {
			output.Info(fmt.Sprintf("Core %s already exists.", core))
			return nil
		}
		output.Info(fmt.Sprintf("Create core %s.", core))
		args := make([]string, 0)
		args = append(args, "-c", s.SolrCoreName(p, core))
		if conf.(map[string]interface{})["conf_dir"] != nil {
			if err := s.solrExtactConfigDir(conf.(map[string]interface{})["conf_dir"].(string)); err != nil {
				return err
			}
			args = append(args, "-d", s.solrGetTempDir())
		}
		if _, err := s.solrCommand("create_core", args...); err != nil {
			return err
		}
		// NEEDS TESTING
		if conf.(map[string]interface{})["core_properties"] != nil {
			corePropPath := filepath.Join(s.DataPath(), s.SolrCoreName(p, core), "core.properties")
			coreProps := fmt.Sprintf("name=%s\n", s.SolrCoreName(p, core)) + conf.(map[string]interface{})["core_properties"].(string)
			if err := ioutil.WriteFile(
				corePropPath, []byte(coreProps), 0755,
			); err != nil {
				return errors.WithStack(errors.WithMessage(err, s.DisplayName()))
			}
		}
	}
	return nil
}

// SolrCoreName returns the name of a given solr core in reference to given project.
func (s *Service) SolrCoreName(p *Project, core string) string {
	if p == nil {
		return core
	}
	return fmt.Sprintf("%s_%s", p.Name, core)
}

// SolrHasCore check if solr already has given core.
func (s *Service) SolrHasCore(p *Project, core string) bool {
	port, _ := s.Port()
	if port == 0 {
		return false
	}
	core = s.SolrCoreName(p, core)
	resp, err := http.Get(fmt.Sprintf(
		"http://localhost:%d/solr/admin/cores?action=STATUS&core=%s",
		port, core,
	))
	if err != nil {
		return false
	}
	rawResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	coreStatusData := make(map[string]interface{})
	if err := json.Unmarshal(rawResp, &coreStatusData); err != nil {
		return false
	}
	return coreStatusData["status"].(map[string]interface{})[core].(map[string]interface{})["name"] != nil

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

// solrPostSetup configures solr for given service definition.
func (s *Service) solrPostSetup(d *def.Service, p *Project) error {
	if !s.IsSolr() {
		return errors.WithStack(errors.WithMessage(ErrServiceNotSolr, s.DisplayName()))
	}
	if err := s.SolrAddConfigSets(d, p); err != nil {
		return err
	}
	if err := s.SolrAddCores(d, p); err != nil {
		return err
	}
	return nil
}

func (s *Service) solrExtactConfigDir(value string) error {
	confDirTar, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return errors.WithStack(err)
	}
	confDirTarReader := bytes.NewReader(confDirTar)
	gzReader, err := gzip.NewReader(confDirTarReader)
	if err != nil {
		return err
	}
	defer gzReader.Close()
	tarReader := tar.NewReader(gzReader)
	os.RemoveAll(s.solrGetTempDir())
	os.MkdirAll(s.solrGetTempDir(), 0755)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil
			}
			return errors.WithStack(err)
		}
		if header == nil {
			continue
		}
		target := filepath.Join(s.solrGetTempDir(), header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			{
				break
			}
		case tar.TypeReg:
			{
				os.MkdirAll(filepath.Dir(target), 0755)
				f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
				if err != nil {
					return errors.WithStack(err)
				}
				if _, err := io.Copy(f, tarReader); err != nil {
					return errors.WithStack(err)
				}
				f.Close()
				break
			}
		}
	}
}

func (s *Service) solrGetTempDir() string {
	return filepath.Join("/tmp", "solrconf")
}
