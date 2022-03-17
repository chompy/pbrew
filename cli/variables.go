package cli

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"

	"gitlab.com/contextualcode/pbrew/core"

	"github.com/spf13/cobra"
)

func loadVars() (def.Variables, string) {
	// get global flag
	global := varCmd.PersistentFlags().Lookup("global").Value.String() == "true"
	// load project
	proj, err := getProject()
	if !os.IsNotExist(err) {
		handleError(err)
	}
	// get var group name
	varGroupName := core.GlobalVariableFile
	if proj != nil && !global {
		varGroupName = proj.Name
	}
	// load vars
	vars, err := core.LoadVariables(varGroupName)
	if errors.Is(err, os.ErrNotExist) {
		return make(def.Variables), varGroupName
	}
	handleError(err)
	return vars, varGroupName
}

var varCmd = &cobra.Command{
	Use:     "variables [-g global]",
	Aliases: []string{"var", "vars", "v"},
	Short:   "Manage variables.",
}

var varSetCmd = &cobra.Command{
	Use:   "set key value",
	Short: "Start variable.",
	Run: func(cmd *cobra.Command, args []string) {
		// validate
		if len(args) == 0 {
			handleError(ErrInvalidArgs)
		}
		// load vars
		vars, varGroup := loadVars()
		// fetch value
		fi, _ := os.Stdin.Stat()
		hasStdin := fi.Mode()&os.ModeDevice == 0
		value := ""
		if len(args) >= 2 {
			value = args[1]
		} else if hasStdin {
			valueBytes, err := ioutil.ReadAll(os.Stdin)
			handleError(err)
			value = string(valueBytes)
		}
		// set
		handleError(vars.Set(strings.TrimSpace(args[0]), value))
		// save
		handleError(core.SaveVariables(varGroup, vars))
	},
}

var varListCmd = &cobra.Command{
	Use:   "list [--json]",
	Short: "List project variable.",
	Run: func(cmd *cobra.Command, args []string) {
		vars, _ := loadVars()
		// json
		if cmd.PersistentFlags().Lookup("json").Value.String() == "true" {
			varsJson, err := json.Marshal(vars)
			handleError(err)
			output.WriteStdout(string(varsJson))
			return
		}
		// table
		tableRows := make([][]string, 0)
		for k := range vars {
			tableRows = append(tableRows, []string{k, vars.GetString(k)})
		}
		sort.Slice(tableRows, func(i, j int) bool {
			return tableRows[i][0] < tableRows[j][0]
		})
		drawTable(
			[]string{"KEY", "VALUE"},
			tableRows,
		)
	},
}

func init() {
	varCmd.PersistentFlags().StringP("name", "n", "", "variable name")
	varCmd.PersistentFlags().BoolP("global", "g", false, "set global variable")
	varListCmd.PersistentFlags().Bool("json", false, "output as json")
	varCmd.AddCommand(varSetCmd)
	varCmd.AddCommand(varListCmd)
	RootCmd.AddCommand(varCmd)
}
