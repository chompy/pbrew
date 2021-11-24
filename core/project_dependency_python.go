package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
)

func (p *Project) depPythonBuildRequirementsTxt(d interface{}) error {
	switch d := d.(type) {
	case *def.App:
		{
			done := output.Duration("Build requirements.txt.")
			deps := d.Dependencies
			for pyVersion, pyDeps := range map[int]map[string]string{2: deps.Python2, 3: deps.Python3} {
				// make dir
				depPath := filepath.Join(p.DepInstallPath(d), fmt.Sprintf("python%d", pyVersion))
				if err := os.Mkdir(depPath, 0755); err != nil {
					if !os.IsExist(err) {
						return err
					}
				}
				// generate requirements
				out := ""
				for name, ver := range pyDeps {
					out += name
					if ver != "*" {
						out += fmt.Sprintf("==%s", ver)
					}
					out += "\n"
				}
				// write file
				reqTxtPath := filepath.Join(depPath, "requirements.txt")
				if err := ioutil.WriteFile(reqTxtPath, []byte(out), 0755); err != nil {
					return err
				}
			}
			done()
		}
	}
	return nil
}

// DepPythonPipInstall runs python pip install for application dependencies.
func (p *Project) DepPythonPipInstall(d interface{}) error {
	switch d := d.(type) {
	case *def.App:
		{
			done := output.Duration(fmt.Sprintf("Install Python dependencies for %s.", d.Name))
			if err := p.depPythonBuildRequirementsTxt(d); err != nil {
				return err
			}
			done2 := output.Duration("Pip install.")
			for pyMajorVer, pyFullVer := range map[int]string{2: "2.7.18", 3: "3.10.0"} {
				pipBinPath := filepath.Join(GetDir(HomeDir), ".pyenv", "versions", pyFullVer, "bin", "pip")
				depPath := filepath.Join(p.DepInstallPath(d), fmt.Sprintf("python%d", pyMajorVer))
				reqTxtPath := filepath.Join(depPath, "requirements.txt")
				cmd := NewShellCommand()
				cmd.Args = []string{
					"--norc", "-c",
					fmt.Sprintf("%s install -r %s --prefix %s", pipBinPath, reqTxtPath, depPath),
				}
				cmd.Env = brewEnv()
				if err := cmd.Interactive(); err != nil {
					return errors.WithStack(err)
				}
			}
			done2()
			done()
		}
	}
	return nil
}
