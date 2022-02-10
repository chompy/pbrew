package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"gitlab.com/contextualcode/pbrew/core"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
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
		handleError(brewService.MySQLDump(database))
	},
}

var databaseListSchemas = &cobra.Command{
	Use:   "list [--json]",
	Short: "List available schemas for current project.",
	Run: func(cmd *cobra.Command, args []string) {
		proj, err := getProject()
		handleError(err)
		serv, err := getService(databaseCmd, proj, []string{"mariadb", "mysql"})
		handleError(err)
		brewService := core.Service{
			BrewName: "mysql",
		}
		brewService.SetDefinition(proj, &serv)
		schemeas := brewService.MySQLGetSchemas()
		// json
		if cmd.PersistentFlags().Lookup("json").Value.String() == "true" {
			schemasJson, err := json.Marshal(schemeas)
			handleError(err)
			output.WriteStdout(string(schemasJson) + "\n")
			return
		}
		out := make([][]string, 0)
		for i := range schemeas {
			out = append(out, []string{schemeas[i]})
		}
		drawTable([]string{"SCHEMA"}, out)
	},
}

func init() {
	databaseCmd.PersistentFlags().StringP("service", "s", "", "name of database service")
	databaseCmd.PersistentFlags().StringP("database", "d", "", "database/schema to use")
	databaseListSchemas.PersistentFlags().Bool("json", false, "output in json")
	databaseCmd.AddCommand(databaseSql)
	databaseCmd.AddCommand(databaseDump)
	databaseCmd.AddCommand(databaseListSchemas)
	RootCmd.AddCommand(databaseCmd)
}
