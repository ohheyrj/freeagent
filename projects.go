package main

import (
	"fmt"
	"net/url"
	"path"
)

var validProjectViews = map[string]bool{
	"active":    true,
	"completed": true,
	"cancelled": true,
	"hidden":    true,
}

func cmdListProjects(c *Client, view string, asJSON bool) error {
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

	projects, err := apiRequestPaged[project](c, endpoint, "projects")
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
		return printJSON(out)
	}

	rows := make([][]string, 0, len(out))
	for _, p := range out {
		rows = append(rows, []string{
			p.ID,
			p.Name,
			p.Status,
		})
	}
	return printTable([]string{"ID", "Name", "Status"}, rows)
}

func fetchProjectNames(c *Client) (map[string]string, error) {
	type project struct {
		URL  string `json:"url"`
		Name string `json:"name"`
	}
	projects, err := apiRequestPaged[project](c, "/projects", "projects")
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(projects))
	for _, p := range projects {
		m[p.URL] = p.Name
	}
	return m, nil
}
