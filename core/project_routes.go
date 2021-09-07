package core

import (
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

// GetUpstreamPort returns port for given app upstream.
func (p *Project) GetUpstreamPort(d interface{}) int {
	portMap, err := LoadPortMap()
	if err != nil {
		output.Warn(err.Error())
		return 0
	}
	switch d := d.(type) {
	case *def.App:
		{
			port, err := portMap.UpstreamPort(d, p)
			if err != nil {
				output.Warn(err.Error())
				return 0
			}
			return port
		}
	}
	return 0
}
