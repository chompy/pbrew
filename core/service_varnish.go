package core

import (
	"bytes"
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
)

// IsVarnish returns true if service is varnish.
func (s *Service) IsVarnish() bool {
	return strings.HasPrefix(s.Name, "varnish")
}

// IsSolrRunning returns true if solr is running.
func (s *Service) IsVarnishRunning() bool {
	c := NewShellCommand()
	c.Command = "pgrep"
	/*p, err := s.Port()
	if err != nil {
		return false
	}*/
	c.Args = []string{"-f", "varnishd"}
	var buf bytes.Buffer
	c.Stdout = &buf
	c.Interactive()
	return strings.TrimSpace(buf.String()) != ""
}

func (s *Service) varnishConfigParams() map[string]interface{} {
	projName := ""
	relationships := make([]map[string]interface{}, 0)
	if s.project != nil {
		projName = s.project.Name
		portMap, err := LoadPortMap()
		if err != nil {
			return nil
		}
		for _, r := range s.project.Apps {
			port, err := portMap.UpstreamPort(r, s.project)
			if err != nil {
				continue
			}
			relationships = append(relationships, map[string]interface{}{
				"Name": r.Name,
				"Port": port,
			})
		}
		// TODO services
	}
	vcl := ""
	switch d := s.definition.(type) {
	case *def.Service:
		{
			vcl = d.Configuration["vcl"].(string)
			break
		}
	default:
		{
			return nil
		}
	}
	return map[string]interface{}{
		"Vcl":      vcl,
		"Services": relationships,
		"Project":  projName,
	}
}
