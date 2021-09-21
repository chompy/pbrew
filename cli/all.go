package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gitlab.com/contextualcode/pbrew/core"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Perform global operations.",
}

var allStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop all services.",
	Run: func(cmd *cobra.Command, args []string) {
		// itterate services and stop
		serviceList, err := core.LoadServiceList()
		handleError(err)
		for _, service := range serviceList {
			if !service.IsRunning() {
				continue
			}
			if err := service.Stop(); err != nil {
				output.Warn(err.Error())
				output.IndentLevel--
				continue
			}
			time.Sleep(time.Second)
		}
		// stop nginx
		nginx := core.NginxService()
		if nginx.IsRunning() {
			if err := nginx.Stop(); err != nil {
				output.Warn(err.Error())
				output.IndentLevel--
			}
		}
		done := output.Duration("Clean up.")
		// delete config directory
		os.RemoveAll(core.GetDir(core.ConfDir))
		// delete run directory
		os.RemoveAll(core.GetDir(core.RunDir))
		done()
	},
}

var allPurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge all pbrew files.",
	Run: func(cmd *cobra.Command, args []string) {
		output.Info("STARTING ALL PURGE IN 5 SECONDS")
		time.Sleep(time.Second * 5)
		// all stop
		allStopCmd.Run(cmd, args)
		// delete dirs
		done := output.Duration("Delete data directories.")
		os.RemoveAll(core.GetDir(core.DataDir))
		os.RemoveAll(core.GetDir(core.VarsDir))
		os.RemoveAll(core.GetDir(core.LogDir))
		os.RemoveAll(core.GetDir(core.MntDir))
		// TODO option to delete homebrew dir?
		done()
	},
}

var allServicesCmd = &cobra.Command{
	Use:   "services [--json]",
	Short: "List all services and their status.",
	Run: func(cmd *cobra.Command, args []string) {
		brewServices, err := core.LoadServiceList()
		handleError(err)
		runningServices, err := core.ProjectTrackServices()
		handleError(err)
		portMaps, err := core.LoadPortMap()
		handleError(err)
		projectTracks, err := core.ProjectTrackGet()
		handleError(err)
		tableRows := make([][]string, 0)
		for _, service := range brewServices {
			// check if already listed
			alreadyCreated := false
			for _, tableRow := range tableRows {
				if tableRow[0] == service.BrewName {
					alreadyCreated = true
					break
				}
			}
			if alreadyCreated {
				continue
			}
			// get status
			status := "not installed"
			if service.IsInstalled() {
				status = "stopped"
			}
			for _, runningService := range runningServices {
				if runningService == service.BrewName {
					status = "running"
					break
				}
			}
			// get port
			port, err := portMaps.ServicePort(service)
			handleError(err)
			// get projects
			projects := make([]string, 0)
			for _, pt := range projectTracks {
				for _, ptService := range pt.Services {
					if ptService == service.BrewName {
						projects = append(projects, pt.Name)
						break
					}
				}
			}
			// create row
			tableRows = append(tableRows, []string{
				service.BrewName,
				fmt.Sprintf("%d", port),
				status,
				strings.Join(projects, ","),
			})
		}
		sort.Slice(tableRows, func(i int, j int) bool {
			return strings.Compare(tableRows[i][0], tableRows[j][0]) < 0
		})
		drawTable(
			[]string{"NAME", "PORT", "STATUS", "PROJECTS"},
			tableRows,
		)
	},
}

func init() {
	allServicesCmd.PersistentFlags().Bool("json", false, "output in json")
	allCmd.AddCommand(allStopCmd)
	allCmd.AddCommand(allPurgeCmd)
	allCmd.AddCommand(allServicesCmd)
	RootCmd.AddCommand(allCmd)
}
