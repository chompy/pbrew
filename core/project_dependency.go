package core

import (
	"path/filepath"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
)

// DepInstallPath returns the path that dependencies should be installed to.
func (p *Project) DepInstallPath(d interface{}) string {
	switch d := d.(type) {
	case *def.App:
		{
			return filepath.Join(d.Path, ".global")
		}
	}
	return filepath.Join(p.Path, ".global")
}
