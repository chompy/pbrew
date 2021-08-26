package core

import (
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

// RunCommand runs a shell command.
func RunCommand(cmdString string) error {
	cmd := exec.Command("sh", "-c", cmdString)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errors.WithStack(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.WithStack(err)
	}
	if err := cmd.Start(); err != nil {
		return errors.WithStack(err)
	}
	go func() {
		io.Copy(os.Stderr, stderr)
	}()
	if _, err := io.Copy(os.Stdout, stdout); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
