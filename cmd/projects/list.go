package projects

import (
	"fmt"
	"github.com/ohheyrj/freeagent/internal/api"
	"github.com/ohheyrj/freeagent/internal/output"
	"net/url"
	"path"

	"github.com/spf13/cobra"
)

var validProjectViews = map[string]bool{
	"active":    true,
	"completed": true,
	"cancelled": true,
	"hidden":    true,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		view, _ := cmd.Flags().GetString("view")
		asJSON, _ := cmd.Flags().GetBool("json")
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		return listProjects(client, view, asJSON)
	},
}

func init() {
	listCmd.Flags().String("view", "", "filter: active|completed|cancelled|hidden")
	listCmd.Flags().Bool("json", false, "output as JSON")
	Cmd.AddCommand(listCmd)
}

func listProjects(c *api.Client, view string, asJSON bool) error {
	if view != "" && !validProjectViews[view] {
		return fmt.Errorf("invalid view %q (must be active, completed, cancelled, or hidden)", view)
	}

	endpoint := "/projects"
	if view != "" {
		q := url.Values{}
		q.Set("view", view)
		endpoint = endpoint + "?" + q.Encode()
	}
	type project struct {
		URL    string `json:"url"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}

	projects, err := api.Paginate[project](c, endpoint, "projects")
	if err != nil {
		return err
	}

	type projectOut struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}

	out := make([]projectOut, 0, len(projects))
	for _, p := range projects {
		out = append(out, projectOut{
			ID:     path.Base(p.URL),
			Name:   p.Name,
			Status: p.Status,
		})
	}

	if asJSON {
		return output.JSON(out)
	}

	rows := make([][]string, 0, len(out))
	for _, p := range out {
		rows = append(rows, []string{
			p.ID,
			p.Name,
			p.Status,
		})
	}
	return output.Table([]string{"ID", "Name", "Status"}, rows)
}
