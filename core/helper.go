package core

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func wildcardCompare(original string, test string) bool {
	test = regexp.QuoteMeta(test)
	test = strings.Replace(test, "\\*", ".*", -1)
	regex, err := regexp.Compile(test + "$")
	if err != nil {
		return false
	}
	return regex.MatchString(original)
}

func appPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", errors.WithStack(err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return filepath.Dir(execPath), nil
}

func userPath() string {
	homePath, err := os.UserHomeDir()
	if err != nil {
		// TODO
		panic(err)
	}
	return filepath.Join(homePath, ".pbrew")
}

func scanPlatformAppYaml(topPath string, disableOverrides bool) [][]string {
	o := make([][]string, 0)
	appYamlPaths := make([]string, 0)
	filepath.Walk(topPath, func(path string, f os.FileInfo, err error) error {
		// check sub directory
		if f.IsDir() && f.Name() != "." && path != topPath {
			for _, appYamlFilename := range appYamlFilenames {
				possiblePath := filepath.Join(path, appYamlFilename)
				if _, err := os.Stat(possiblePath); !os.IsNotExist(err) {
					appYamlPaths = append(appYamlPaths, possiblePath)
				}
			}
			return filepath.SkipDir
		}
		// check root directory
		for _, appYamlFilename := range appYamlFilenames {
			if f.Name() == appYamlFilename {
				appYamlPaths = append(appYamlPaths, path)
			}
		}
		return nil
	})
	for _, appYamlFilename := range appYamlFilenames {
		for _, appYamlPath := range appYamlPaths {
			if strings.HasSuffix(appYamlPath, appYamlFilename) {
				hasOut := false
				for i := range o {
					if filepath.Dir(o[i][0]) == filepath.Dir(appYamlPath) {
						if !disableOverrides {
							o[i] = append(o[i], appYamlPath)
						}
						hasOut = true
					}
				}
				if !hasOut {
					oo := make([]string, 1)
					oo[0] = appYamlPath
					o = append(o, oo)
				}
			}
		}
	}
	return o
}

func loadYAML(name string, out interface{}) error {
	appPath, err := appPath()
	if err != nil {
		return errors.WithStack(err)
	}
	yamlRaw, err := ioutil.ReadFile(filepath.Join(appPath, "conf", name+".yaml"))
	if err != nil {
		return errors.WithStack(err)
	}
	if err := yaml.Unmarshal(yamlRaw, out); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
