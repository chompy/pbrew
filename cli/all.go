package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	Use:   "services [--json]",
	Short: "List all services and their status.",
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
			ports := make([]string, 0)
			for _, port := range status.Ports {
				ports = append(ports, fmt.Sprintf("%d", port))
			}
			tableRows = append(tableRows, []string{
				status.Name,
				strings.Join(ports, ","),
				status.Status,
				strings.Join(status.Projects, ","),
			})
		}
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
