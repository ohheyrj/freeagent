package api

func FetchTaskNames(c *Client) (map[string]string, error) {
	type task struct {
		URL  string `json:"url"`
		Name string `json:"name"`
	}
	tasks, err := Paginate[task](c, "/tasks", "tasks")
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(tasks))
	for _, t := range tasks {
		m[t.URL] = t.Name
	}
	return m, nil
}
