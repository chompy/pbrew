package core

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

const (
	BrewDir = iota
	RunDir
	ConfDir
	DataDir
	VarsDir
	MntDir
	HomeDir
	LogDir
	AppDir
	UserDir
	BottleDir
)

var appDirectories = map[int]string{
	BrewDir:   filepath.Join(getUserPath(), "homebrew"),
	RunDir:    filepath.Join(getUserPath(), "run"),
	ConfDir:   filepath.Join(getUserPath(), "conf"),
	DataDir:   filepath.Join(getUserPath(), "data"),
	VarsDir:   filepath.Join(getUserPath(), "vars"),
	MntDir:    filepath.Join(getUserPath(), "mnt"),
	HomeDir:   filepath.Join(getUserPath(), "home"),
	LogDir:    filepath.Join(getUserPath(), "logs"),
	AppDir:    getAppPath(),
	UserDir:   getUserPath(),
	BottleDir: filepath.Join(getUserPath(), "bottles"),
}

// GetDir returns given key's path.
func GetDir(key int) string {
	return appDirectories[key]
}

// InitDirs creates directories needed by app.
func InitDirs() error {
	for _, dir := range appDirectories {
		if err := os.MkdirAll(dir, mkdirPerm); err != nil {
			if errors.Is(err, os.ErrExist) {
				continue
			}
			return errors.WithStack(err)
		}
	}
	return nil
}

func resolveUserPath(path string) string {
	if strings.HasPrefix(path, "~") {
		homePath, err := os.UserHomeDir()
		if err != nil {
			output.Warn(err.Error())
			return path
		}
		path = filepath.Join(homePath, path[1:])
	}
	return path
}

func getAppPath() string {
	execPath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		panic(err)
	}
	return filepath.Dir(execPath)
}

func getUserPath() string {
	conf, err := LoadConfig()
	if err != nil {
		output.Warn(err.Error())
	}
	return resolveUserPath(conf.UserDir)
}
