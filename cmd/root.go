package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ohheyrj/freeagent/cmd/auth"
	"github.com/ohheyrj/freeagent/cmd/projects"
	"github.com/ohheyrj/freeagent/cmd/tasks"
	"github.com/ohheyrj/freeagent/cmd/timeslips"
	"github.com/ohheyrj/freeagent/cmd/users"
)

// version is set at build time via -ldflags "-X github.com/ohheyrj/freeagent/cmd.version=..."
var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "freeagent",
	Short:   "FreeAgent CLI",
	Version: version,
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
