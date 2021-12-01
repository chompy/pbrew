package core

import (
	"fmt"
	"sort"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

// BrewInstallAll installs all services used by pbrew.
func BrewInstallAll(reinstall bool) error {
	phpExtList, err := LoadPHPExtensionList()
	if err != nil {
		return err
	}
	serviceList, err := LoadServiceList()
	if err != nil {
		return err
	}
	keys := make([]string, 0)
	for key := range serviceList {
		if key[0] == '_' {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, name := range keys {
		def := serviceList[name]
		done := output.Duration(fmt.Sprintf("Install '%s.'", name))
		if def.IsInstalled() {
			if !reinstall {
				output.Info("Already installed.")
				done()
				continue
			}
			if err := def.Uninstall(); err != nil {
				return err
			}
		}
		if err := def.InstallDependencies(); err != nil {
			return err
		}
		if err := def.Install(); err != nil {
			return err
		}
		if def.IsPHP() {
			for _, ext := range phpExtList {
				if err := def.PHPInstallExtension(ext); err != nil {
					return err
				}
			}
		}
		done()
	}

	return nil
}
