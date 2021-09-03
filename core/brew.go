package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"github.com/pkg/errors"
)

const brewPath = ".pbrew/homebrew"

// BrewPath is the path the homebrew install.
func BrewPath() string {
	homePath, err := os.UserHomeDir()
	if err != nil {
		// TODO
		panic(err)
	}
	return filepath.Join(homePath, brewPath)
}

// IsBrewInstalled returns true if homebrew is installed.
func IsBrewInstalled() bool {
	if _, err := os.Stat(BrewPath()); os.IsNotExist(err) {
		return false
	}
	return true
}

// BrewInstall installs homebrew in application root.
func BrewInstall() error {
	done := output.Duration("Install Homebrew.")
	if err := os.MkdirAll(BrewPath(), 0755); err != nil {
		return errors.WithStack(err)
	}
	cmd := NewShellCommand()
	cmd.Args = []string{"-c", fmt.Sprintf(brewInstall, BrewPath())}
	if err := cmd.Interactive(); err != nil {
		return errors.WithStack(err)
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
	binPath := filepath.Join(BrewPath(), "bin/brew")
	cmd := NewShellCommand()
	cmd.Args = []string{"-c", binPath + " " + strings.Join(subCmds, " ")}
	if err := cmd.Interactive(); err != nil {
		return errors.WithStack(err)
	}
	done()
	return nil
}
