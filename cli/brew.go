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
		// get project
		proj, err := getProject()
		handleError(err)
		// get service
		service := brewCmdSelectService(proj)
		if service == nil {
			handleError(ErrServiceNotFound)
		}
		serviceList, err := core.LoadServiceList()
		handleError(err)
		serviceDef, err := serviceList.MatchDef(service)
		handleError(err)
		handleError(serviceDef.Compile())
	},
}

func brewCmdSelectService(proj *core.Project) interface{} {
	serviceName := brewCmd.PersistentFlags().Lookup("service").Value.String()
	var service interface{} = proj.Apps[0]
	if serviceName != "" {
		for _, sapp := range proj.Apps {
			if sapp.Name == serviceName {
				service = sapp
				break
			}
		}
		for _, sservice := range proj.Services {
			if sservice.Name == serviceName {
				service = sservice
				break
			}
		}
	}
	return service
}

func init() {
	brewCmd.PersistentFlags().StringP("service", "s", "", "name of service")
	brewCmd.AddCommand(brewCompileCmd)
	RootCmd.AddCommand(brewCmd)
}
