package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
)

func (p *Project) depPHPBuildComposerJSON(d interface{}) error {
	switch d := d.(type) {
	case *def.App:
		{
			done := output.Duration("Build composer.json.")
			deps := d.Dependencies
			if deps.PHP.Require["composer/composer"] == "" {
				if deps.PHP.Require == nil {
					deps.PHP.Require = make(map[string]string)
				}
				deps.PHP.Require["composer/composer"] = "^1"
			}
			composerJSON, err := json.Marshal(deps.PHP)
			if err != nil {
				return err
			}
			if err := os.Mkdir(p.DepInstallPath(d), 0755); err != nil {
				if !os.IsExist(err) {
					return err
				}
			}
			os.Remove(filepath.Join(p.DepInstallPath(d), "composer.lock"))
			if err := ioutil.WriteFile(
				filepath.Join(p.DepInstallPath(d), "composer.json"),
				composerJSON,
				0655,
			); err != nil {
				return err
			}
			done()
		}
	}
	return nil
}

// DepPHPComposerInstall runs composer install for application dependencies.
func (p *Project) DepPHPComposerInstall(d interface{}) error {
	switch d := d.(type) {
	case *def.App:
		{
			done := output.Duration(fmt.Sprintf("Install PHP dependencies for %s.", d.Name))
			if err := p.depPHPBuildComposerJSON(d); err != nil {
				return err
			}
			done2 := output.Duration("Composer install.")
			serviceList, err := LoadServiceList()
			if err != nil {
				return err
			}
			brewService, err := serviceList.MatchDef(d)
			if err != nil {
				return err
			}
			phpBinPath := filepath.Join(GetDir(BrewDir), "opt", brewService.BrewAppName(), "bin", "php")
			composerBinPath := filepath.Join(GetDir(BrewDir), "bin", "composer")
			if err := p.Command(d, fmt.Sprintf(
				"%s %s install -d %s",
				phpBinPath, composerBinPath, p.DepInstallPath(d),
			)); err != nil {
				return err
			}
			done2()
			done()
		}
	}
	return nil
}
