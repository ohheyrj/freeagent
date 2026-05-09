package main

import (
	"net/url"
	"path"
)

type listTimeslipOpts struct {
	View      string
	UserID    string
	ProjectID string
	TaskID    string
	FromDate  string
	ToDate    string
}

func cmdListTimeslips(c *Client, opts listTimeslipOpts, asJSON bool) error {
	q := url.Values{}
	if opts.View != "" {
		q.Set("view", opts.View)
	}
	if opts.UserID != "" {
		q.Set("user", baseURL+"/users/"+opts.UserID)
	}
	if opts.ProjectID != "" {
		q.Set("project", baseURL+"/projects/"+opts.ProjectID)
	}
	if opts.TaskID != "" {
		q.Set("task", baseURL+"/tasks/"+opts.TaskID)
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

	timeslips, err := apiRequestPaged[timeslip](c, endpoint, "timeslips")
	if err != nil {
		return err
	}

	projectNames, err := fetchProjectNames(c)
	if err != nil {
		return err
	}
	userNames, err := fetchUserNames(c)
	if err != nil {
		return err
	}
	taskNames, err := fetchTaskNames(c)
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
		return printJSON(out)
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
	return printTable([]string{"ID", "User ID", "User Name", "Project ID", "Project Name", "Task ID", "Task Name", "Dated On", "Hours"}, rows)
}
