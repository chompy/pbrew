package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
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
			cmd := NewShellCommand()
			npmBinPath := filepath.Join(GetDir(BrewDir), "bin", "npm")
			nodeModulesPath := filepath.Join(p.DepInstallPath(d), "node_modules")
			cmd.Args = []string{
				"--norc", "-c",
				fmt.Sprintf("%s install %s -g --prefix %s", npmBinPath, p.DepInstallPath(d), nodeModulesPath),
			}
			cmd.Env = brewEnv()
			if err := cmd.Interactive(); err != nil {
				return errors.WithStack(err)
			}
			done2()
			done()
		}
	}
	return nil
}
