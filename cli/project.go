package cli

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gitlab.com/contextualcode/pbrew/core"
)

var projectCmd = &cobra.Command{
	Use:     "project",
	Aliases: []string{"proj", "p"},
	Short:   "Manage projects.",
}

var projectStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start project.",
	Run: func(cmd *cobra.Command, args []string) {
		// start project
		proj, err := getProject()
		handleError(err)
		handleError(proj.Start())
		// generate nginx
		handleError(core.NginxAdd(proj))
		// start/reload nginx
		nginx := core.NginxService()
		if nginx == nil {
			handleError(errors.WithMessage(core.ErrServiceNotFound, "nginx"))
		}
		if nginx.IsRunning() {
			handleError(nginx.Reload())
			return
		}
		if !nginx.IsInstalled() {
			handleError(nginx.Install())
		}
		handleError(nginx.Start())
	},
}

var projectStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop project.",
	Run: func(cmd *cobra.Command, args []string) {
		// stop project
		proj, err := getProject()
		handleError(err)
		handleError(proj.Stop())
		// generate nginx
		handleError(core.NginxDel(proj))
		// reload nginx
		nginx := core.NginxService()
		if nginx == nil {
			handleError(errors.WithMessage(core.ErrServiceNotFound, "nginx"))
		}
		if nginx.IsRunning() {
			handleError(nginx.Reload())
			return
		}
		// TODO stop if last project
	},
}

func init() {
	projectCmd.AddCommand(projectStartCmd)
	projectCmd.AddCommand(projectStopCmd)
	RootCmd.AddCommand(projectCmd)
}
