package cli

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gitlab.com/contextualcode/pbrew/core"
)

var projectCmd = &cobra.Command{
	Use:     "project",
	Aliases: []string{"proj", "p"},
	Short:   "Manage projects.",
}

var projectStartCmd = &cobra.Command{
	Use:   "start [--no-mounts] [--no-bottles]",
	Short: "Start project.",
	Run: func(cmd *cobra.Command, args []string) {
		// start project
		proj, err := getProject()
		handleError(err)
		proj.NoMounts = cmd.PersistentFlags().Lookup("no-mounts").Value.String() == "true"
		proj.NoBottles = cmd.PersistentFlags().Lookup("no-bottles").Value.String() == "true"
		handleError(proj.Start())
		// generate nginx
		handleError(core.NginxAdd(proj))
		// start/reload nginx
		nginx := core.NginxService()
		if nginx == nil {
			handleError(errors.WithMessage(core.ErrServiceNotFound, "nginx"))
		}
		if nginx.IsRunning() {
			handleError(nginx.Reload())
			return
		}
		if !nginx.IsInstalled() {
			handleError(nginx.Install())
		}
		handleError(nginx.PreStart(nil, nil))
		handleError(nginx.Start())
	},
}

var projectStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop project.",
	Run: func(cmd *cobra.Command, args []string) {
		// stop project
		proj, err := getProject()
		handleError(err)
		handleError(proj.Stop())
		// generate nginx
		handleError(core.NginxDel(proj))
		// reload nginx
		nginx := core.NginxService()
		if nginx == nil {
			handleError(errors.WithMessage(core.ErrServiceNotFound, "nginx"))
		}
		if nginx.IsRunning() {
			projectTracks, err := core.ProjectTrackGet()
			handleError(err)
			if len(projectTracks) == 0 {
				handleError(nginx.Stop())
				return
			}
			handleError(nginx.Reload())
			return
		}
	},
}

var projectPurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge project.",
	Run: func(cmd *cobra.Command, args []string) {
		// stop project
		projectStopCmd.Run(cmd, args)
		// itterate defs and purge services
		proj, err := getProject()
		handleError(err)
		brewServiceList, err := core.LoadServiceList()
		handleError(err)
		for _, service := range proj.Services {
			brewService, err := brewServiceList.MatchDef(service)
			if err != nil && errors.Is(err, core.ErrServiceNotFound) {
				continue
			}
			handleError(err)
			handleError(brewService.Purge(&service, proj))
		}
		for _, service := range proj.Apps {
			brewService, err := brewServiceList.MatchDef(service)
			if err != nil && errors.Is(err, core.ErrServiceNotFound) {
				continue
			}
			handleError(err)
			handleError(brewService.Purge(service, proj))
		}
	},
}

func init() {
	projectStartCmd.PersistentFlags().Bool("no-mounts", false, "disable symlink mounts")
	projectStartCmd.PersistentFlags().Bool("no-bottles", false, "disable pbrew provided bottles")
	projectCmd.AddCommand(projectStartCmd)
	projectCmd.AddCommand(projectStopCmd)
	projectCmd.AddCommand(projectPurgeCmd)
	RootCmd.AddCommand(projectCmd)
}
