package main

import (
	"path"
)

func cmdListUsers(c *Client, asJSON bool) error {
	type user struct {
		URL       string `json:"url"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	users, err := apiRequestPaged[user](c, "/users", "users")
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
		return printJSON(out)
	}

	rows := make([][]string, 0, len(out))
	for _, u := range out {
		rows = append(rows, []string{
			u.ID,
			u.FirstName,
			u.LastName,
		})
	}
	return printTable([]string{"ID", "First Name", "Last Name"}, rows)
}

func fetchUserNames(c *Client) (map[string]string, error) {
	type user struct {
		URL       string `json:"url"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	users, err := apiRequestPaged[user](c, "/users", "users")
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(users))
	for _, u := range users {
		m[u.URL] = u.FirstName + " " + u.LastName
	}
	return m, nil
}
