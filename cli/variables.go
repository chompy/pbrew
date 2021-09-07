package cli

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

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

func init() {
	varCmd.PersistentFlags().StringP("name", "n", "", "variable name")
	varCmd.PersistentFlags().BoolP("global", "g", false, "set global variable")
	varCmd.AddCommand(varSetCmd)
	RootCmd.AddCommand(varCmd)
}
