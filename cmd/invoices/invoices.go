package invoices

import (
	"github.com/ohheyrj/freeagent/internal/api"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "invoices",
	Short: "Manage invoices",
}

func init() {
	Cmd.AddCommand(newCreateCommand(api.NewClient))
}
