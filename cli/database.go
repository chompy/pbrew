package cli

import (
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/contextualcode/pbrew/core"
)

var databaseCmd = &cobra.Command{
	Use:     "database [-s service] [-d database]",
	Aliases: []string{"mysql", "mariadb", "db"},
	Short:   "Manage database services.",
}

var databaseSql = &cobra.Command{
	Use:   "sql",
	Short: "Access SQL shell.",
	Run: func(cmd *cobra.Command, args []string) {
		proj, err := getProject()
		handleError(err)
		serv, err := getService(databaseCmd, proj, []string{"mariadb", "mysql"})
		handleError(err)
		brewServiceList, err := core.LoadServiceList()
		handleError(err)
		brewService, err := brewServiceList.MatchDef(serv)
		handleError(err)
		database := databaseCmd.PersistentFlags().Lookup("database").Value.String()
		database = proj.ResolveDatabase(database)
		handleError(brewService.MySQLShell(database))
	},
}

var databaseDump = &cobra.Command{
	Use:   "dump",
	Short: "Dump SQL database.",
	Run: func(cmd *cobra.Command, args []string) {
		proj, err := getProject()
		handleError(err)
		serv, err := getService(databaseCmd, proj, []string{"mariadb", "mysql"})
		handleError(err)
		brewServiceList, err := core.LoadServiceList()
		handleError(err)
		brewService, err := brewServiceList.MatchDef(serv)
		handleError(err)
		database := databaseCmd.PersistentFlags().Lookup("database").Value.String()
		database = proj.ResolveDatabase(database)
		handleError(brewService.MySQLDump(database, os.Stdout))
	},
}

func init() {
	databaseCmd.PersistentFlags().StringP("service", "s", "", "name of database service")
	databaseCmd.PersistentFlags().StringP("database", "d", "", "database/schema to use")
	databaseCmd.AddCommand(databaseSql)
	databaseCmd.AddCommand(databaseDump)
	RootCmd.AddCommand(databaseCmd)
}
