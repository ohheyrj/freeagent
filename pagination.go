package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

func apiRequestPaged[T any](c *Client, basePath, jsonKey string) ([]T, error) {
	const perPage = 100

	u, err := url.Parse(basePath)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("per_page", strconv.Itoa(perPage))
	q.Set("page", "1")
	u.RawQuery = q.Encode()

	res, body, err := c.doRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	var first map[string][]T
	if err := json.Unmarshal(body, &first); err != nil {
		return nil, err
	}

	total, _ := strconv.Atoi(res.Header.Get("X-Total-Count"))
	if total == 0 {
		return first[jsonKey], nil
	}

	all := make([]T, 0, total)
	all = append(all, first[jsonKey]...)

	pages := (total + perPage - 1) / perPage
	for p := 2; p <= pages; p++ {
		q.Set("page", strconv.Itoa(p))
		u.RawQuery = q.Encode()

		var pageN map[string][]T
		if err := c.apiRequest(http.MethodGet, u.String(), nil, &pageN); err != nil {
			return nil, err
		}
		all = append(all, pageN[jsonKey]...)
	}
	return all, nil
}
