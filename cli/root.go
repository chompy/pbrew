package cli

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

// RootCmd is the top level command.
var RootCmd = &cobra.Command{
	Use:     "pbrew [-v verbose]",
	Version: "DEV",
	Run: func(cmd *cobra.Command, args []string) {
		commandIntro(cmd.Version)
		output.WriteStdout("\nAvailable Commands:\n")
		displayCommandList(cmd)
	},
}

// Execute - run root command
func Execute() error {
	// hack that allows old style semicolon (:) seperated
	// subcommands to work
	args := make([]string, 1)
	args[0] = os.Args[0]
	if len(os.Args) > 1 {
		args = append(args, strings.Split(os.Args[1], ":")...)
		args = append(args, os.Args[2:]...)
	}
	os.Args = args
	return RootCmd.Execute()
}

func init() {
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "show more verbose output")
}
