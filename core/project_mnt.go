package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"github.com/pkg/errors"
)

// SetupMounts sets up symlinks for mount directories.
func (p *Project) SetupMounts() error {
	if p.NoMounts {
		output.LogInfo("No mounts flag enabled, skipping mount setup.")
		return nil
	}
	done := output.Duration("Setup mounts.")
	mntPath := filepath.Join(userPath(), mntDir, p.Name)
	destPaths := make([]string, 0)
	for _, app := range p.Apps {
		for dest, mount := range app.Mounts {
			// build path to destination directory inside app root
			destPath := filepath.Join(p.Path, strings.Trim(dest, string(filepath.Separator)))
			destPath = strings.TrimRight(strings.ReplaceAll(
				destPath, ":", "_",
			), string(filepath.Separator))
			// check if dest path has already been mounted
			alreadyHasDest := false
			for _, existingDestPaths := range destPaths {
				if destPath == existingDestPaths {
					alreadyHasDest = true
					break
				}
			}
			if alreadyHasDest {
				continue
			}
			destPaths = append(destPaths, destPath)
			// build source path inside user dir
			srcPath := filepath.Join(mntPath, strings.Trim(mount.SourcePath, string(filepath.Separator)))
			srcPath = strings.ReplaceAll(
				srcPath, ":", "_",
			)
			output.LogInfo(fmt.Sprintf("Mount %s to %s.", srcPath, destPath))
			if err := os.MkdirAll(srcPath, mkdirPerm); err != nil {
				if !errors.Is(err, os.ErrExist) {
					return errors.WithStack(err)
				}
			}

			if err := os.RemoveAll(destPath); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return errors.WithStack(err)
				}
			}
			if err := os.Symlink(srcPath, destPath); err != nil {
				return errors.WithStack(err)
			}
		}
	}
	done()
	return nil
}
