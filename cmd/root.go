package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ohheyrj/freeagent/cmd/auth"
	"github.com/ohheyrj/freeagent/cmd/projects"
	"github.com/ohheyrj/freeagent/cmd/tasks"
	"github.com/ohheyrj/freeagent/cmd/timeslips"
	"github.com/ohheyrj/freeagent/cmd/users"
)

var rootCmd = &cobra.Command{
	Use:     "freeagent",
	Short:   "FreeAgent CLI",
	Version: "0.1.0",
}

func init() {
	rootCmd.AddCommand(auth.Cmd)
	rootCmd.AddCommand(projects.Cmd)
	rootCmd.AddCommand(tasks.Cmd)
	rootCmd.AddCommand(users.Cmd)
	rootCmd.AddCommand(timeslips.Cmd)
}

func Execute() error {
	return rootCmd.Execute()
}
