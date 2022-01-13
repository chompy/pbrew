package core

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

// IsSolr returns true if service is php.
func (s *Service) IsSolr() bool {
	return strings.HasPrefix(s.BrewAppName(), "java") && strings.Contains(s.StartCmd, "solr")
}

// IsSolrRunning returns true if SOLR is running.
func (s *Service) IsSolrRunning() bool {
	var buf bytes.Buffer
	cmd := NewShellCommand()
	cmd.Stdout = &buf
	cmd.Env = []string{
		"JAVA_HOME=" + filepath.Join(GetDir(BrewDir), "opt", "java11"),
	}
	cmd.Command = filepath.Join(GetDir(BrewDir), "opt", "solr", "bin", "solr")
	port, _ := s.Port()
	cmd.Args = []string{"status", "-p", fmt.Sprintf("%d", port)}
	if err := cmd.Interactive(); err != nil {
		output.Warn(err.Error())
		return false
	}
	return strings.Contains(buf.String(), fmt.Sprintf("port %d", port))
}
