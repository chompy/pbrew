package cli

import (
	"github.com/spf13/cobra"
	"gitlab.com/contextualcode/pbrew/core"
)

var routerCmd = &cobra.Command{
	Use:     "router",
	Aliases: []string{"r"},
	Short:   "Manage router.",
}

var routerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start nginx router.",
	Run: func(cmd *cobra.Command, args []string) {
		nginx := core.NginxService()
		if nginx == nil {
			handleError(core.ErrServiceNotFound)
		}
		if !nginx.IsInstalled() {
			handleError(nginx.Install())
		}
		handleError(nginx.Start())
	},
}

var routerStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop nginx router.",
	Run: func(cmd *cobra.Command, args []string) {
		nginx := core.NginxService()
		if nginx == nil {
			handleError(core.ErrServiceNotFound)
		}
		handleError(nginx.Stop())
	},
}

var routerAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add project to router.",
	Run: func(cmd *cobra.Command, args []string) {
		proj, err := getProject()
		handleError(err)
		handleError(core.NginxAdd(proj))
	},
}

func init() {
	routerCmd.AddCommand(routerStartCmd)
	routerCmd.AddCommand(routerStopCmd)
	routerCmd.AddCommand(routerAddCmd)
	RootCmd.AddCommand(routerCmd)
}
