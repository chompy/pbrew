package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

func (p *Project) getAppShellCommand(d *def.App) (ShellCommand, error) {
	// get app brew service
	serviceList, err := LoadServiceList()
	if err != nil {
		return ShellCommand{}, err
	}
	brewAppService, err := serviceList.MatchDef(d)
	if err != nil {
		return ShellCommand{}, err
	}
	if !brewAppService.IsRunning() {
		return ShellCommand{}, errors.WithStack(errors.WithMessage(ErrServiceNotRunning, brewAppService.DisplayName()))
	}
	brewServiceList := make([]*Service, 0)
	brewServiceList = append(brewServiceList, brewAppService)
	for _, service := range p.Services {
		brewService, err := serviceList.MatchDef(&service)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				continue
			}
			return ShellCommand{}, err
		}
		brewServiceList = append(brewServiceList, brewService)
	}
	// generate pathes
	envPaths := []string{
		filepath.Join(p.Path, ".global", "bin"),
		filepath.Join(p.Path, ".global", "vendor", "bin"),
		filepath.Join(p.Path, ".global", "node_modules", "bin"),
		filepath.Join(p.Path, ".platformsh", "bin"),
		filepath.Join(GetDir(HomeDir), ".pyenv", "versions", "3.10.0", "bin"),
		filepath.Join(GetDir(HomeDir), ".pyenv", "versions", "2.7.18", "bin"),
	}
	// inject env vars
	env := make([]string, 0)
	env = append(env, ServicesEnv(brewServiceList)...)
	for k, v := range env {
		if strings.HasPrefix(v, "PATH") {
			env[k] = "PATH=" + strings.Join(envPaths, ":") + ":" + strings.TrimPrefix(env[k], "PATH=")
		}
	}
	//env = append(env, "HOME="+p.Path)
	//env = append(env, fmt.Sprintf("NVM_DIR=%s/.nvm", GetDir(HomeDir)))
	env = append(env, fmt.Sprintf("TERM=%s", os.Getenv("TERM")))
	for k, v := range p.Env(d) {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	// run interactive shell
	cmd := NewShellCommand()
	cmd.Args = []string{}
	cmd.Env = env
	return cmd, nil
}

// Shell opens a shell in given app context.
func (p *Project) Shell(d *def.App) error {
	output.Info(fmt.Sprintf("Access shell for %s.", d.Name))
	cmd, err := p.getAppShellCommand(d)
	if err != nil {
		return err
	}
	if err := cmd.Drop(); err != nil {
		return err
	}
	return nil
}

// Command executes a shell command in given app context.
func (p *Project) Command(d *def.App, cmdStr string) error {
	output.LogInfo(fmt.Sprintf("Run command '%s' in '%s'.", cmdStr, d.Name))
	cmd, err := p.getAppShellCommand(d)
	if err != nil {
		return err
	}
	cmdStr = "source $(brew --prefix nvm)/nvm.sh && " + cmdStr
	cmd.Args = []string{"-c", cmdStr}
	if err := cmd.Interactive(); err != nil {
		return err
	}
	return nil
}
