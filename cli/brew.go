package cli

import (
	"github.com/spf13/cobra"
	"gitlab.com/contextualcode/pbrew/core"
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
		handleError(err)
		handleError(serviceDef.Compile())
	},
}

func init() {
	brewCmd.PersistentFlags().StringP("service", "s", "", "name of service")
	brewCmd.AddCommand(brewCompileCmd)
	RootCmd.AddCommand(brewCmd)
}
