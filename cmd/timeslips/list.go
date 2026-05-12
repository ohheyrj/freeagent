package timeslips

import (
	"fmt"
	"net/url"
	"path"

	"github.com/spf13/cobra"

	"freeagent/internal/api"
	"freeagent/internal/output"
)

type ListTimeslipOpts struct {
	View      string
	UserID    string
	ProjectID string
	TaskID    string
	FromDate  string
	ToDate    string
}

var validTimeslipViews = map[string]bool{
	"all":      true,
	"unbilled": true,
	"running":  true,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List timeslips",
	RunE: func(cmd *cobra.Command, args []string) error {
		var opts ListTimeslipOpts
		opts.View, _ = cmd.Flags().GetString("view")
		opts.UserID, _ = cmd.Flags().GetString("user")
		opts.ProjectID, _ = cmd.Flags().GetString("project")
		opts.TaskID, _ = cmd.Flags().GetString("task")
		opts.FromDate, _ = cmd.Flags().GetString("from")
		opts.ToDate, _ = cmd.Flags().GetString("to")
		asJSON, _ := cmd.Flags().GetBool("json")
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		return ListTimeslips(client, opts, asJSON)
	},
}

func init() {
	listCmd.Flags().String("view", "", "filter: all|unbilled|running")
	listCmd.Flags().String("user", "", "filter by user ID")
	listCmd.Flags().String("project", "", "filter by project ID")
	listCmd.Flags().String("task", "", "filter by task ID")
	listCmd.Flags().String("from", "", "from date (YYYY-MM-DD)")
	listCmd.Flags().String("to", "", "to date (YYYY-MM-DD)")
	listCmd.Flags().Bool("json", false, "output as JSON")
	Cmd.AddCommand(listCmd)
}

func ListTimeslips(c *api.Client, opts ListTimeslipOpts, asJSON bool) error {
	if opts.View != "" && !validTimeslipViews[opts.View] {
		return fmt.Errorf("invalid view %q (must be all, running or unbilled)", opts.View)
	}
	q := url.Values{}
	if opts.View != "" {
		q.Set("view", opts.View)
	}
	if opts.UserID != "" {
		q.Set("user", api.BaseURL+"/users/"+opts.UserID)
	}
	if opts.ProjectID != "" {
		q.Set("project", api.BaseURL+"/projects/"+opts.ProjectID)
	}
	if opts.TaskID != "" {
		q.Set("task", api.BaseURL+"/tasks/"+opts.TaskID)
	}
	if opts.FromDate != "" {
		q.Set("from_date", opts.FromDate)
	}
	if opts.ToDate != "" {
		q.Set("to_date", opts.ToDate)
	}

	endpoint := "/timeslips"
	if encoded := q.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	type timeslip struct {
		URL       string `json:"url"`
		User      string `json:"user"`
		ProjectID string `json:"project"`
		Task      string `json:"task"`
		DatedOn   string `json:"dated_on"`
		Hours     string `json:"hours"`
	}

	timeslips, err := api.Paginate[timeslip](c, endpoint, "timeslips")
	if err != nil {
		return err
	}

	projectNames, err := api.FetchProjectNames(c)
	if err != nil {
		return err
	}
	userNames, err := api.FetchUserNames(c)
	if err != nil {
		return err
	}
	taskNames, err := api.FetchTaskNames(c)
	if err != nil {
		return err
	}

	type timeslipOut struct {
		ID          string `json:"id"`
		UserID      string `json:"user_id"`
		UserName    string `json:"user_name"`
		ProjectID   string `json:"project_id"`
		ProjectName string `json:"project_name"`
		TaskID      string `json:"task_id"`
		TaskName    string `json:"task_name"`
		DatedOn     string `json:"dated_on"`
		Hours       string `json:"hours"`
	}

	out := make([]timeslipOut, 0, len(timeslips))
	for _, t := range timeslips {
		out = append(out, timeslipOut{
			ID:          path.Base(t.URL),
			UserID:      path.Base(t.User),
			UserName:    userNames[t.User],
			ProjectID:   path.Base(t.ProjectID),
			ProjectName: projectNames[t.ProjectID],
			TaskID:      path.Base(t.Task),
			TaskName:    taskNames[t.Task],
			DatedOn:     t.DatedOn,
			Hours:       t.Hours,
		})
	}

	if asJSON {
		return output.JSON(out)
	}

	rows := make([][]string, 0, len(out))
	for _, t := range out {
		rows = append(rows, []string{
			t.ID,
			t.UserID,
			t.UserName,
			t.ProjectID,
			t.ProjectName,
			t.TaskID,
			t.TaskName,
			t.DatedOn,
			t.Hours,
		})
	}
	return output.Table([]string{"ID", "User ID", "User Name", "Project ID", "Project Name", "Task ID", "Task Name", "Dated On", "Hours"}, rows)
}
