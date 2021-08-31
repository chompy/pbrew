package cli

import (
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
		// start nginx
		nginx := core.NginxService()
		if nginx == nil {
			handleError(core.ErrServiceNotFound)
		}
		handleError(nginx.Start())
	},
}

func init() {
	projectCmd.AddCommand(projectStartCmd)
	RootCmd.AddCommand(projectCmd)
}
