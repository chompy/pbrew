package core

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
)

type ShellCommand struct {
	Command string
	Args    []string
	Env     []string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

func NewShellCommand() ShellCommand {
	return ShellCommand{
		Command: filepath.Join(GetDir(BrewDir), "bin", "zsh"),
		Args:    []string{},
		Env:     []string{},
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}
}

// Interactive creates an interactive passthru shell.
func (s ShellCommand) Interactive() error {
	cmd := exec.Command(s.Command, s.Args...)
	cmd.Stderr = s.Stderr
	cmd.Stdout = s.Stdout
	cmd.Stdin = s.Stdin
	cmd.Env = s.Env
	//io.WriteString(os.Stdout, "=== INTERACTIVE SHELL =====================\n")
	if err := cmd.Run(); err != nil {
		return errors.WithStack(err)
	}
	//io.WriteString(os.Stdout, "===========================================\n")
	return nil
}

// Drop executes the shell command by dropping to the system shell and exiting pbrew.
func (s ShellCommand) Drop() error {
	args := make([]string, 0)
	args = append(args, filepath.Base(s.Command))
	args = append(args, s.Args...)
	if err := syscall.Exec(s.Command, args, s.Env); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
