package api

func FetchUserNames(c *Client) (map[string]string, error) {
	type user struct {
		URL       string `json:"url"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	users, err := Paginate[user](c, "/users", "users")
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(users))
	for _, u := range users {
		m[u.URL] = u.FirstName + " " + u.LastName
	}
	return m, nil
}
