package cli

import (
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gitlab.com/contextualcode/pbrew/core"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
	"golang.org/x/term"
)

// handleError handles an error.
func handleError(err error) {
	output.Verbose = checkFlag(RootCmd, "verbose")
	output.Error(err)
}

// checkFlag returns true if given flag is set.
func checkFlag(cmd *cobra.Command, name string) bool {
	if cmd == nil {
		return false
	}
	flag := cmd.Flags().Lookup(name)
	return flag != nil && flag.Value.String() != "false"
}

// getProject fetches the project at the current working directory.
func getProject() (*core.Project, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	proj, err := core.LoadProject(cwd)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return proj, err
}

// getService fetches a service definition.
func getService(cmd *cobra.Command, proj *core.Project, filterType []string) (def.Service, error) {
	name := cmd.PersistentFlags().Lookup("service").Value.String()
	for _, serv := range proj.Services {
		for _, t := range filterType {
			if (serv.Name == name || name == "") && t == serv.GetTypeName() {
				return serv, nil
			}
		}
	}
	if name == "" {
		return def.Service{}, errors.WithStack(errors.WithMessage(ErrServiceNotFound, strings.Join(filterType, ",")))
	}
	return def.Service{}, errors.WithStack(errors.WithMessage(ErrServiceNotFound, name))
}

// drawTable draws an ASCII table to stdout.
func drawTable(head []string, data [][]string) {
	if len(data) == 0 {
		output.WriteStdout("=== NO DATA ===\n")
		return
	}
	w, _, _ := term.GetSize(int(os.Stdin.Fd()))
	if w == 0 {
		w = 256
	}
	w -= (len(head) * 4)
	truncateString := func(size int, value string) string {
		if len(value) <= size {
			return value
		}
		if size <= 4 {
			return string(value[0]) + "..."
		}
		return value[0:size-3] + "..."
	}
	// calculate column widths
	oColWidths := make([]int, len(data[0]))
	for i, sv := range head {
		if len(sv) > oColWidths[i] {
			oColWidths[i] = len(sv)
		}
	}
	for _, v := range data {
		for i, sv := range v {
			if len(sv) > oColWidths[i] {
				oColWidths[i] = len(sv)
			}
		}
	}
	// calculate max display width per colume
	mColWidths := make([]int, len(oColWidths))
	for i := range mColWidths {
		mColWidths[i] = w / len(oColWidths)
	}
	for i := range mColWidths {
		ni := i + 1
		if ni >= len(mColWidths) {
			ni = 0
		}
		if oColWidths[i] < mColWidths[i] {
			diff := mColWidths[i] - oColWidths[i]
			mColWidths[i] -= diff
			mColWidths[ni] += diff
		}
	}
	// truncate head cols
	for i := range head {
		head[i] = truncateString(mColWidths[i], head[i])
	}
	// truncate data cols
	for i := range data {
		for j := range data[i] {
			data[i][j] = truncateString(mColWidths[j], data[i][j])
		}
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(head)
	table.SetAutoWrapText(true)
	table.SetBorder(false)
	table.AppendBulk(data)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	output.WriteStdout("\n")
	table.Render()
	output.WriteStdout("\n")
}

// commandIntro displays introduction information about pbrew.
func commandIntro(version string) {
	output.WriteStdout(output.Color(strings.Repeat("=", 32), 32) + "\n")
	output.WriteStdout(" PBREW BY CONTEXTUAL CODE \n")
	output.WriteStdout(output.Color(strings.Repeat("=", 32), 32) + "\n")
	output.WriteStdout(output.Color(" VERSION "+version, 35) + "\n")
}
