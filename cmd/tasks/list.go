package tasks

import (
	"fmt"
	"net/url"
	"path"
	"strconv"

	"github.com/spf13/cobra"

	"freeagent/internal/api"
	"freeagent/internal/output"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, _ := cmd.Flags().GetString("project")
		asJSON, _ := cmd.Flags().GetBool("json")
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		return listTasks(client, projectID, asJSON)
	},
}

func init() {
	listCmd.Flags().String("project", "", "filter by project ID")
	listCmd.Flags().Bool("json", false, "output as JSON")
	Cmd.AddCommand(listCmd)
}

func listTasks(c *api.Client, projectID string, asJSON bool) error {
	endpoint := "/tasks"
	if projectID != "" {
		if _, err := strconv.Atoi(projectID); err != nil {
			return fmt.Errorf("invalid project ID %q: must be numeric", projectID)
		}
		q := url.Values{}
		q.Set("project", api.BaseURL+"/projects/"+projectID)
		endpoint += "?" + q.Encode()
	}
	type task struct {
		URL       string `json:"url"`
		Name      string `json:"name"`
		Status    string `json:"status"`
		ProjectID string `json:"project"`
	}

	tasks, err := api.Paginate[task](c, endpoint, "tasks")
	if err != nil {
		return err
	}

	projectNames, err := api.FetchProjectNames(c)
	if err != nil {
		return err
	}

	type taskOut struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Status      string `json:"status"`
		ProjectID   string `json:"project_id"`
		ProjectName string `json:"project_name"`
	}

	out := make([]taskOut, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, taskOut{
			ID:          path.Base(t.URL),
			Name:        t.Name,
			Status:      t.Status,
			ProjectID:   path.Base(t.ProjectID),
			ProjectName: projectNames[t.ProjectID],
		})
	}

	if asJSON {
		return output.JSON(out)
	}

	rows := make([][]string, 0, len(out))
	for _, t := range out {
		rows = append(rows, []string{
			t.ID,
			t.Name,
			t.Status,
			t.ProjectName,
			t.ProjectID,
		})
	}
	return output.Table([]string{"ID", "Name", "Status", "Project Name", "Project ID"}, rows)
}
