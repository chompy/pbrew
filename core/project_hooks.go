package core

import (
	"fmt"
	"path/filepath"
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
)

func (p *Project) hookCmdReplace(cmd string) string {
	cmd = strings.ReplaceAll(cmd, "/app", p.Path)
	cmd = strings.ReplaceAll(cmd, "/usr/bin", filepath.Join(GetDir(BrewDir), "bin"))
	return strings.TrimSpace(cmd)
}

func (p *Project) executeHookCmd(cmdStr string, d *def.App) error {
	cmdStr = p.hookCmdReplace(cmdStr)
	if cmdStr == "" {
		return nil
	}
	cmd, err := p.getAppShellCommand(d)
	if err != nil {
		return err
	}
	cmd.Args = []string{
		"--init-file", filepath.Join(GetDir(HomeDir), ".bash_profile"), "-c", cmdStr,
	}
	if err := cmd.Interactive(); err != nil {
		return err
	}
	return nil
}

// Build executes build hooks for given app.
func (p *Project) Build(d *def.App) error {
	done := output.Duration(fmt.Sprintf("Execute build hook for %s.", d.Name))
	if err := p.executeHookCmd(d.Hooks.Build, d); err != nil {
		return err
	}
	done()
	return nil
}

// Deploy executes deploy hooks for given app.
func (p *Project) Deploy(d *def.App) error {
	done := output.Duration(fmt.Sprintf("Execute deploy hook for %s.", d.Name))
	if err := p.executeHookCmd(d.Hooks.Deploy, d); err != nil {
		return err
	}
	done()
	return nil
}

// PostDeploy executes post deploy hooks for given app.
func (p *Project) PostDeploy(d *def.App) error {
	done := output.Duration(fmt.Sprintf("Execute post deploy hook for %s.", d.Name))
	if err := p.executeHookCmd(d.Hooks.PostDeploy, d); err != nil {
		return err
	}
	done()
	return nil
}
