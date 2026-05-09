package main

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
)

func cmdListTasks(c *Client, projectID string, asJSON bool) error {
	endpoint := "/tasks"
	if projectID != "" {
		if _, err := strconv.Atoi(projectID); err != nil {
			return fmt.Errorf("invalid project ID %q: must be numeric", projectID)
		}
		q := url.Values{}
		q.Set("project", baseURL+"/projects/"+projectID)
		endpoint += "?" + q.Encode()
	}
	type task struct {
		URL       string `json:"url"`
		Name      string `json:"name"`
		Status    string `json:"status"`
		ProjectID string `json:"project"`
	}

	tasks, err := apiRequestPaged[task](c, endpoint, "tasks")
	if err != nil {
		return err
	}

	projectNames, err := fetchProjectNames(c)
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
		return printJSON(out)
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
	return printTable([]string{"ID", "Name", "Status", "Project Name", "Project ID"}, rows)
}

func fetchTaskNames(c *Client) (map[string]string, error) {
	type task struct {
		URL  string `json:"url"`
		Name string `json:"name"`
	}
	tasks, err := apiRequestPaged[task](c, "/tasks", "tasks")
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(tasks))
	for _, t := range tasks {
		m[t.URL] = t.Name
	}
	return m, nil
}
