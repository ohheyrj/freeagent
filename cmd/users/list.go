package users

import (
	"path"

	"github.com/spf13/cobra"

	"freeagent/internal/api"
	"freeagent/internal/output"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List users",
	RunE: func(cmd *cobra.Command, args []string) error {
		asJSON, _ := cmd.Flags().GetBool("json")
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		return ListUsers(client, asJSON)
	},
}

func init() {
	listCmd.Flags().Bool("json", false, "output as JSON")
	Cmd.AddCommand(listCmd)
}

func ListUsers(c *api.Client, asJSON bool) error {
	type user struct {
		URL       string `json:"url"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	users, err := api.Paginate[user](c, "/users", "users")
	if err != nil {
		return err
	}

	type userOut struct {
		ID        string `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	out := make([]userOut, 0, len(users))
	for _, u := range users {
		out = append(out, userOut{
			ID:        path.Base(u.URL),
			FirstName: u.FirstName,
			LastName:  u.LastName,
		})
	}

	if asJSON {
		return output.JSON(out)
	}

	rows := make([][]string, 0, len(out))
	for _, u := range out {
		rows = append(rows, []string{
			u.ID,
			u.FirstName,
			u.LastName,
		})
	}
	return output.Table([]string{"ID", "First Name", "Last Name"}, rows)
}
