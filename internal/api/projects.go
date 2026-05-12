package api

func FetchProjectNames(c *Client) (map[string]string, error) {
	type project struct {
		URL  string `json:"url"`
		Name string `json:"name"`
	}
	projects, err := Paginate[project](c, "/projects", "projects")
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(projects))
	for _, p := range projects {
		m[p.URL] = p.Name
	}
	return m, nil
}
