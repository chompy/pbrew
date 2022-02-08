package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gitlab.com/contextualcode/pbrew/core"
)

var routerCmd = &cobra.Command{
	Use:     "router",
	Aliases: []string{"r", "nginx"},
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
		handleError(nginx.PreStart())
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

var routerListCmd = &cobra.Command{
	Use:   "list [--json]",
	Short: "List routes for project.",
	Run: func(cmd *cobra.Command, args []string) {
		proj, err := getProject()
		handleError(err)
		hostNames := core.GetHostNames(proj.Routes)
		// json
		if cmd.PersistentFlags().Lookup("json").Value.String() == "true" {
			out := make(map[string][]def.Route)
			for _, hostName := range hostNames {
				out[hostName] = core.GetRoutesForHostName(hostName, proj.Routes)
			}
			jsonOut, err := json.Marshal(out)
			handleError(err)
			output.WriteStdout(string(jsonOut))
			return
		}
		// table
		tableRows := make([][]string, 0)
		for _, hostName := range hostNames {
			routes := core.GetRoutesForHostName(hostName, proj.Routes)
			upstreams := make([]string, 0)
			for _, route := range routes {
				if route.Type != "upstream" || route.Upstream == "" {
					continue
				}
				hasUpstream := false
				for _, upstream := range upstreams {
					if route.Upstream == upstream {
						hasUpstream = true
						break
					}
				}
				if !hasUpstream {
					upstreams = append(upstreams, route.Upstream)
				}
			}
			tableRows = append(tableRows, []string{
				core.ProjectDefaultHostName(proj, hostName),
				fmt.Sprintf("%d", len(routes)),
				strings.Join(upstreams, ","),
			})
		}
		drawTable(
			[]string{"HOST", "NUMBER OF ROUTES", "UPSTREAMS"},
			tableRows,
		)

	},
}

func init() {
	routerListCmd.PersistentFlags().Bool("json", false, "output as json")
	routerCmd.AddCommand(routerStartCmd)
	routerCmd.AddCommand(routerStopCmd)
	routerCmd.AddCommand(routerAddCmd)
	routerCmd.AddCommand(routerDelCmd)
	routerCmd.AddCommand(routerListCmd)
	RootCmd.AddCommand(routerCmd)
}
