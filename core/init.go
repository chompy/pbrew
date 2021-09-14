package core

import (
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

const mkdirPerm = 0755

// InitApp runs first time initalization procedures.
func InitApp() error {
	done := output.Duration("Pbrew init.")
	// init dirs
	if err := InitDirs(); err != nil {
		return err
	}
	// install brew
	if !IsBrewInstalled() {
		if err := BrewInstall(); err != nil {
			return err
		}
	}
	done()
	return nil
}
