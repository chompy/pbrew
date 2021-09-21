package cli

import (
	"os"
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

func init() {
	allCmd.AddCommand(allStopCmd)
	allCmd.AddCommand(allPurgeCmd)
	RootCmd.AddCommand(allCmd)
}
