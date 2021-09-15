package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"github.com/pkg/errors"
)

// IsBrewInstalled returns true if homebrew is installed.
func IsBrewInstalled() bool {
	if _, err := os.Stat(GetDir(BrewDir)); os.IsNotExist(err) {
		return false
	}
	return true
}

// BrewInstall installs homebrew in application root.
func BrewInstall() error {
	done := output.Duration("Install Homebrew.")
	cmd := NewShellCommand()
	cmd.Args = []string{"-c", fmt.Sprintf(brewInstall, GetDir(BrewDir))}
	cmd.Env = os.Environ()
	if err := cmd.Interactive(); err != nil {
		return err
	}
	// taps
	if err := brewCommand("tap", "shivammathur/php"); err != nil {
		return err
	}
	// dependencies
	if err := brewCommand("install", "openssl"); err != nil {
		return err
	}
	done()
	return nil
}

// brewCommand runs a homebrew command.
func brewCommand(subCmds ...string) error {
	if !IsBrewInstalled() {
		return errors.WithStack(ErrBrewNotInstalled)
	}
	done := output.Duration("Run brew " + strings.Join(subCmds, " ") + ".")
	binPath := filepath.Join(GetDir(BrewDir), "bin/brew")
	cmd := NewShellCommand()
	cmd.Args = []string{"-c", binPath + " " + strings.Join(subCmds, " ")}
	cmd.Env = os.Environ()
	if err := cmd.Interactive(); err != nil {
		return errors.WithStack(err)
	}
	done()
	return nil
}
