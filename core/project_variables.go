package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
)

// GlobalVariableFile is the name of the global variable file.
const GlobalVariableFile = "_global"

func variablePath(name string) string {
	return filepath.Join(userPath(), "vars", name)
}

func LoadVariables(name string) (def.Variables, error) {
	output.LogInfo(fmt.Sprintf("Load '%s' variables.", name))
	raw, err := ioutil.ReadFile(variablePath(name))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	vars := make(def.Variables)
	if err := json.Unmarshal(raw, &vars); err != nil {
		return nil, errors.WithStack(err)
	}
	return vars, nil
}

func SaveVariables(name string, vars def.Variables) error {
	output.LogInfo(fmt.Sprintf("Save '%s' variables.", name))
	out, err := json.Marshal(vars)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := ioutil.WriteFile(variablePath(name), out, 0755); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Variables returns variables for project, merged with global.
func (p *Project) Variables(d interface{}) (def.Variables, error) {
	out := make(def.Variables)
	// add def vars
	switch d := d.(type) {
	case *def.App:
		{
			out.Merge(d.Variables)
			break
		}
	case *def.AppWorker:
		{
			out.Merge(d.Variables)
			break
		}
	}
	// add global vars
	globalVars, err := LoadVariables(GlobalVariableFile)
	if err != nil {
		return nil, err
	}
	out.Merge(globalVars)
	// add project vars
	projVars, err := LoadVariables(p.Name)
	if err != nil {
		return nil, err
	}
	out.Merge(projVars)
	return out, nil
}
