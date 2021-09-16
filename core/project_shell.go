package core

import (
	"fmt"

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
		return ShellCommand{}, errors.WithStack(errors.WithMessage(ErrServiceNotRunning, brewAppService.BrewName))
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
	// inject env vars
	env := make([]string, 0)
	env = append(env, ServicesEnv(brewServiceList)...)
	//env = append(env, "HOME="+p.Path)
	env = append(env, fmt.Sprintf("PS1=%s-%s> ", p.Name, d.Name))
	for k, v := range p.Env(d) {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	// run interactive shell
	cmd := NewShellCommand()
	cmd.Args = []string{"--norc"}
	cmd.Env = env
	return cmd, nil
}

// Shell opens a shell for given app.
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
