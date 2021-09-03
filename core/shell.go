package core

import (
	"io"
	"log"
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
		Command: "/bin/bash",
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
	if err := cmd.Run(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Drop executes the shell command by dropping to the system shell and exiting pbrew.
func (s ShellCommand) Drop() error {
	args := make([]string, 0)
	args = append(args, filepath.Base(s.Command))
	args = append(args, s.Args...)

	log.Println(s.Command, args)
	if err := syscall.Exec(s.Command, args, s.Env); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
