package core

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"github.com/pkg/errors"
)

// IsBrewInstalled returns true if homebrew is installed.
func IsBrewInstalled() bool {
	if _, err := os.Stat(filepath.Join(GetDir(BrewDir), "bin")); os.IsNotExist(err) {
		return false
	}
	return true
}

// BrewInstall installs homebrew in application root.
func BrewInstall() error {
	done := output.Duration("Install Homebrew.")
	cmd := NewShellCommand()
	cmd.Args = []string{"--norc", "-c", fmt.Sprintf(brewInstall, GetDir(BrewDir))}
	cmd.Env = brewEnv()
	if err := cmd.Interactive(); err != nil {
		return err
	}
	done()
	return BrewInit()
}

// BrewInit runs initialization commands on Homebrew environment.
func BrewInit() error {
	done := output.Duration("Initialize Homebrew environment.")
	// tap shivammathur/php
	if err := brewCommand("tap", "shivammathur/php"); err != nil {
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
	cmd.Args = []string{"--norc", "-c", binPath + " " + strings.Join(subCmds, " ")}
	cmd.Env = brewEnv()
	if err := cmd.Interactive(); err != nil {
		return errors.WithStack(err)
	}
	done()
	return nil
}

func brewEnv() []string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	return []string{
		fmt.Sprintf("HOMEBREW_CELLAR=%s", filepath.Join(GetDir(BrewDir), "Cellar")),
		fmt.Sprintf("HOMEBREW_PREFIX=%s", GetDir(BrewDir)),
		fmt.Sprintf("HOMEBREW_REPOSITORY=%s", GetDir(BrewDir)),
		fmt.Sprintf("HOMEBREW_SHELLENV_PREFIX=%s", GetDir(BrewDir)),
		fmt.Sprintf("LD_LIBRARY_PATH=%s", filepath.Join(GetDir(BrewDir), "lib/gcc/11")),
		fmt.Sprintf("HOME=%s", GetDir(HomeDir)),
		fmt.Sprintf("USER=%s", user.Username),
		fmt.Sprintf("PATH=%s:/bin:/usr/bin:/usr/sbin", filepath.Join(GetDir(BrewDir), "bin")),
		fmt.Sprintf("CPATH=%s", filepath.Join(GetDir(BrewDir), "include")),
		fmt.Sprintf("NVM_DIR=%s/.nvm", GetDir(HomeDir)),
		"RUBY_CFLAGS=-DUSE_FFI_CLOSURE_ALLOC",
	}
}

func brewInfo(name string) (map[string]interface{}, error) {
	binPath := filepath.Join(GetDir(BrewDir), "bin/brew")
	out, err := exec.Command(binPath, "info", name, "--json").Output()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	info := make([]map[string]interface{}, 0)
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, errors.WithStack(err)
	}
	if len(info) == 0 {
		return nil, errors.WithStack(errors.WithMessage(ErrServiceNotFound, name))
	}
	return info[0], nil
}

func brewAppName(name string) string {
	brewAppNamePath := strings.Split(strings.Trim(name, "/"), "/")
	return brewAppNamePath[len(brewAppNamePath)-1]
}
