package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

func (p *Project) depNodeBuildPackageJSON(d interface{}) error {
	switch d := d.(type) {
	case *def.App:
		{
			done := output.Duration("Build package.json.")
			deps := d.Dependencies
			packageJSON, err := json.Marshal(map[string]interface{}{
				"name":         d.Name,
				"dependencies": deps.NodeJS,
			})
			if err != nil {
				return err
			}
			if err := os.Mkdir(p.DepInstallPath(d), 0755); err != nil {
				if !os.IsExist(err) {
					return err
				}
			}
			if err := ioutil.WriteFile(
				filepath.Join(p.DepInstallPath(d), "package.json"),
				packageJSON,
				0655,
			); err != nil {
				return err
			}
			done()
		}
	}
	return nil
}

// DepNodeNpmInstall runs composer install for application dependencies.
func (p *Project) DepNodeNpmInstall(d interface{}) error {
	switch d := d.(type) {
	case *def.App:
		{
			done := output.Duration(fmt.Sprintf("Install Node dependencies for %s.", d.Name))
			if err := p.depNodeBuildPackageJSON(d); err != nil {
				return err
			}
			done2 := output.Duration("Npm install.")
			if err := p.Command(d, fmt.Sprintf(
				"source $(brew --prefix nvm)/nvm.sh && npm install %s --prefix %s",
				p.DepInstallPath(d), p.DepInstallPath(d),
			)); err != nil {
				return err
			}
			nodeBinDir := filepath.Join(p.DepInstallPath(d), "node_modules", "bin")
			os.RemoveAll(nodeBinDir)
			os.MkdirAll(nodeBinDir, mkdirPerm)
			if err := p.Command(d, fmt.Sprintf(
				"cd %s && ln -s ../*/bin/* .",
				nodeBinDir,
			)); err != nil {
				return err
			}
			done2()
			done()
		}
	}
	return nil
}
