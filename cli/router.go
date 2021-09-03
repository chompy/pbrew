package cli

import (
	"github.com/pkg/errors"
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
			handleError(errors.WithMessage(core.ErrServiceNotFound, "nginx"))
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
			handleError(errors.WithMessage(core.ErrServiceNotFound, "nginx"))
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
		nginx := core.NginxService()
		if nginx == nil {
			handleError(errors.WithMessage(core.ErrServiceNotFound, "nginx"))
		}
		handleError(nginx.Reload())
	},
}

var routerDelCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"del"},
	Short:   "Delete project from router.",
	Run: func(cmd *cobra.Command, args []string) {
		proj, err := getProject()
		handleError(err)
		handleError(core.NginxDel(proj))
		nginx := core.NginxService()
		if nginx == nil {
			handleError(errors.WithMessage(core.ErrServiceNotFound, "nginx"))
		}
		handleError(nginx.Reload())
	},
}

func init() {
	routerCmd.AddCommand(routerStartCmd)
	routerCmd.AddCommand(routerStopCmd)
	routerCmd.AddCommand(routerAddCmd)
	routerCmd.AddCommand(routerDelCmd)
	RootCmd.AddCommand(routerCmd)
}
