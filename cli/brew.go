package cli

import (
	"time"

	"github.com/spf13/cobra"
	"gitlab.com/contextualcode/pbrew/core"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

var brewCmd = &cobra.Command{
	Use:     "brew [-s service]",
	Aliases: []string{"homebrew", "hb"},
	Short:   "Manage Homebrew packages.",
}

var brewCompileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile and bottle given service.",
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := brewCmd.PersistentFlags().Lookup("service").Value.String()
		serviceList, err := core.LoadServiceList()
		handleError(err)
		serviceDef, err := serviceList.Match(serviceName)
		if err != nil || serviceDef.Empty() {
			serviceDef = core.Service{
				BrewName: serviceName,
			}
		}
		handleError(serviceDef.Compile())
	},
}

var brewInitCmd = &cobra.Command{
	Use:   "init",
	Short: "(Re)init Homebrew enviroment.",
	Run: func(cmd *cobra.Command, args []string) {
		handleError(core.BrewInit())
	},
}

var brewInstallAllCmd = &cobra.Command{
	Use:   "install-all [--reinstall]",
	Short: "Install every Homebrew service needed by Pbrew.",
	Run: func(cmd *cobra.Command, args []string) {
		output.Info("!! THIS WILL TAKE A LONG TIME, MAKE SURE YOUR COMPUTER DOESN'T GO TO SLEEP. !!")
		time.Sleep(time.Second * 3)
		doReinstall := cmd.PersistentFlags().Lookup("reinstall").Value.String() == "true"
		handleError(core.BrewInstallAll(doReinstall))
	},
}

func init() {
	brewCmd.PersistentFlags().StringP("service", "s", "", "name of service")
	brewInstallAllCmd.PersistentFlags().Bool("reinstall", false, "forces reinstall of all services")
	brewCmd.AddCommand(brewCompileCmd)
	brewCmd.AddCommand(brewInitCmd)
	brewCmd.AddCommand(brewInstallAllCmd)
	RootCmd.AddCommand(brewCmd)
}
