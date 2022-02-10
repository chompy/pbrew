package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gitlab.com/contextualcode/pbrew/core"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

var projectCmd = &cobra.Command{
	Use:     "project",
	Aliases: []string{"proj", "p"},
	Short:   "Manage projects.",
}

var projectStartCmd = &cobra.Command{
	Use:   "start [--no-mounts] [-b use-pbrew-bottles]",
	Short: "Start project.",
	Run: func(cmd *cobra.Command, args []string) {
		// start project
		proj, err := getProject()
		handleError(err)
		proj.NoMounts = cmd.PersistentFlags().Lookup("no-mounts").Value.String() == "true"
		proj.UsePbrewBottles = cmd.PersistentFlags().Lookup("use-pbrew-bottles").Value.String() == "true"
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
		handleError(nginx.PreStart())
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
			projectTracks, err := core.ProjectTrackGet()
			handleError(err)
			if len(projectTracks) == 0 {
				handleError(nginx.Stop())
				return
			}
			handleError(nginx.Reload())
			return
		}
	},
}

var projectPurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge project.",
	Run: func(cmd *cobra.Command, args []string) {
		// stop project
		projectStopCmd.Run(cmd, args)
		// itterate defs and purge services
		proj, err := getProject()
		handleError(err)
		brewServiceList, err := core.LoadServiceList()
		handleError(err)
		for _, service := range proj.Services {
			brewService, err := brewServiceList.MatchDef(service)
			if err != nil && errors.Is(err, core.ErrServiceNotFound) {
				continue
			}
			brewService.SetDefinition(proj, &service)
			handleError(err)
			handleError(brewService.Purge())
		}
		for _, service := range proj.Apps {
			brewService, err := brewServiceList.MatchDef(service)
			if err != nil && errors.Is(err, core.ErrServiceNotFound) {
				continue
			}
			brewService.SetDefinition(proj, service)
			handleError(err)
			handleError(brewService.Purge())
		}
	},
}

var projectStatusCmd = &cobra.Command{
	Use:   "status [--json]",
	Short: "Display current project status.",
	Run: func(cmd *cobra.Command, args []string) {
		proj, err := getProject()
		handleError(err)
		/*projectTracks, err := core.ProjectTrackGet()
		handleError(err)
		projectTrack := core.ProjectTrack{}
		for _, v := range projectTracks {
			if v.Name == proj.Name {
				projectTrack = v
				break
			}
		}*/
		serviceStatues, err := core.GetServiceStatuses()
		handleError(err)
		brewServices, err := proj.GetBrewServices()
		handleError(err)
		out := make([]core.ServiceStatus, 0)
		for _, brewService := range brewServices {
			for _, serviceStatus := range serviceStatues {
				if serviceStatus.Name == brewService.BrewAppName() || serviceStatus.Name == brewService.Name {
					out = append(out, serviceStatus)
					break
				}
			}
		}
		// json
		if cmd.PersistentFlags().Lookup("json").Value.String() == "true" {
			outJson, err := json.Marshal(out)
			handleError(err)
			output.WriteStdout(string(outJson) + "\n")
			return
		}
		rows := make([][]string, 0)
		for _, v := range out {
			ports := make([]string, 0)
			for _, port := range v.Ports {
				ports = append(ports, fmt.Sprintf("%d", port))
			}
			rows = append(rows, []string{
				v.DisplayName,
				v.Name,
				v.Status,
				strings.Join(ports, ","),
			})
		}
		output.WriteStdout("\n >> " + proj.Name + "\n")
		drawTable(
			[]string{"NAME", "BREW NAME", "STATUS", "PORT"},
			rows,
		)
	},
}

func init() {
	projectStartCmd.PersistentFlags().Bool("no-mounts", false, "disable symlink mounts")
	projectStartCmd.PersistentFlags().BoolP("use-pbrew-bottles", "b", false, "enables use of pbrew provided bottles")
	projectStatusCmd.PersistentFlags().Bool("json", false, "output in json")
	projectCmd.AddCommand(projectStartCmd)
	projectCmd.AddCommand(projectStopCmd)
	projectCmd.AddCommand(projectPurgeCmd)
	projectCmd.AddCommand(projectStatusCmd)
	RootCmd.AddCommand(projectCmd)
}
