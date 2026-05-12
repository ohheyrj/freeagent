package timeslips

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"freeagent/internal/api"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a timeslip",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if _, err := strconv.Atoi(id); err != nil {
			return fmt.Errorf("invalid timeslip ID %q: must be numeric", id)
		}
		skipConfirm, _ := cmd.Flags().GetBool("yes")
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		return deleteTimeslip(client, id, skipConfirm)
	},
}

func init() {
	deleteCmd.Flags().BoolP("yes", "y", false, "skip confirmation prompt")
	Cmd.AddCommand(deleteCmd)
}

func deleteTimeslip(c *api.Client, id string, skipConfirm bool) error {
	if !skipConfirm {
		fmt.Fprintf(os.Stderr, "Delete timeslip %s? [y/N] ", id)
		reader := bufio.NewReader(os.Stdin)
		resp, _ := reader.ReadString('\n')
		if !strings.EqualFold(strings.TrimSpace(resp), "y") {
			fmt.Fprintln(os.Stderr, "Cancelled.")
			return nil
		}
	}

	endpoint := "/timeslips/" + id
	if err := c.APIRequest(http.MethodDelete, endpoint, nil, nil); err != nil {
		return err
	}

	fmt.Printf("Deleted timeslip %s\n", id)
	return nil
}
