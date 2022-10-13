package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
		projectTracks, err := core.ProjectTrackGet()
		handleError(err)
		for _, projTrack := range projectTracks {
			for _, ptServiceName := range projTrack.Services {
				service, err := core.ProjectTrackGetService(ptServiceName)
				handleError(err)
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
		// delete project track file
		os.Remove(filepath.Join(core.GetDir(core.UserDir), core.ProjectTrackFile))
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
	Use:     "services [--json]",
	Short:   "List all running services.",
	Aliases: []string{"status"},
	Run: func(cmd *cobra.Command, args []string) {
		statuses, err := core.GetServiceStatuses()
		handleError(err)
		// json out
		if cmd.PersistentFlags().Lookup("json").Value.String() == "true" {
			jsonOut, err := json.Marshal(statuses)
			handleError(err)
			output.WriteStdout(string(jsonOut))
			return
		}
		// table out
		tableRows := make([][]string, 0)
		for _, status := range statuses {
			// create row
			tableRows = append(tableRows, []string{
				status.Project, status.DefName, status.DefType, status.InstanceName, status.Status, fmt.Sprintf("%d", status.Port),
			})
		}
		drawTable(
			[]string{"PROJECT", "NAME", "TYPE", "ID", "STATUS", "PORT"},
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
