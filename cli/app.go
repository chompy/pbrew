package cli

import (
	"github.com/spf13/cobra"
	"gitlab.com/contextualcode/pbrew/core"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
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
		app := appCmdSelectApp(proj)
		handleError(proj.Shell(app))
	},
}

var appBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Run build hook for application.",
	Run: func(cmd *cobra.Command, args []string) {
		proj, err := getProject()
		handleError(err)
		app := appCmdSelectApp(proj)
		handleError(proj.Build(app))
	},
}

var appDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Run Deploy hook for application.",
	Run: func(cmd *cobra.Command, args []string) {
		proj, err := getProject()
		handleError(err)
		app := appCmdSelectApp(proj)
		handleError(proj.Deploy(app))
	},
}

var appPostDeployCmd = &cobra.Command{
	Use:   "post-deploy",
	Short: "Run Deploy hook for application.",
	Run: func(cmd *cobra.Command, args []string) {
		proj, err := getProject()
		handleError(err)
		app := appCmdSelectApp(proj)
		handleError(proj.PostDeploy(app))
	},
}

func appCmdSelectApp(proj *core.Project) *def.App {
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
	return app
}

func init() {
	appCmd.PersistentFlags().StringP("service", "s", "", "name of application")
	appCmd.AddCommand(appShellCmd)
	appCmd.AddCommand(appBuildCmd)
	appCmd.AddCommand(appDeployCmd)
	appCmd.AddCommand(appPostDeployCmd)
	RootCmd.AddCommand(appCmd)
}
