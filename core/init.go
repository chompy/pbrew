package core

import (
	"os"
	"path/filepath"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"github.com/pkg/errors"
)

const mkdirPerm = 0755

// InitApp runs first time initalization procedures.
func InitApp() error {
	done := output.Duration("Pbrew init.")
	// make directories
	if err := os.MkdirAll(userPath(), mkdirPerm); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return errors.WithStack(err)
		}
	}
	if err := os.Mkdir(filepath.Join(userPath(), runDir), mkdirPerm); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return errors.WithStack(err)
		}
	}
	if err := os.Mkdir(filepath.Join(userPath(), varsDir), mkdirPerm); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return errors.WithStack(err)
		}
	}
	if err := os.Mkdir(filepath.Join(userPath(), confDir), mkdirPerm); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return errors.WithStack(err)
		}
	}
	if err := os.Mkdir(filepath.Join(userPath(), dataDir), mkdirPerm); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return errors.WithStack(err)
		}
	}
	if err := os.Mkdir(filepath.Join(userPath(), mntDir), mkdirPerm); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return errors.WithStack(err)
		}
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
