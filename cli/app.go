package cli

import (
	"github.com/spf13/cobra"
)

var appCmd = &cobra.Command{
	Use:     "application [-s service]",
	Aliases: []string{"app", "a"},
	Short:   "Manage applications.",
}

var appShellCmd = &cobra.Command{
	Use:     "shell",
	Aliases: []string{"sh"},
	Short:   "Create shell for application.",
	Run: func(cmd *cobra.Command, args []string) {
		proj, err := getProject()
		handleError(err)
		appName := appCmd.PersistentFlags().Lookup("service").Value.String()
		app := proj.Apps[0]
		if appName != "" {
			for _, sapp := range proj.Apps {
				if sapp.Name == appName {
					app = sapp
					break
				}
			}
		}
		handleError(proj.Shell(app))
	},
}

func init() {
	appCmd.PersistentFlags().StringP("service", "s", "", "name of application")
	appCmd.AddCommand(appShellCmd)
	RootCmd.AddCommand(appCmd)
}
