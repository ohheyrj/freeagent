package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

const baseURL = "https://api.freeagent.com/v2"

type Client struct {
	HTTPClient   *http.Client
	AccessToken  string
	RefreshToken string
	ClientID     string
	ClientSecret string
}

func newFaClient() (*Client, error) {
	c := &Client{
		HTTPClient:   http.DefaultClient,
		AccessToken:  os.Getenv("FREEAGENT_ACCESS_TOKEN"),
		RefreshToken: os.Getenv("FREEAGENT_REFRESH_TOKEN"),
		ClientID:     os.Getenv("FREEAGENT_CLIENT_ID"),
		ClientSecret: os.Getenv("FREEAGENT_CLIENT_SECRET"),
	}
	var missing []string
	if c.AccessToken == "" {
		missing = append(missing, "FREEAGENT_ACCESS_TOKEN")
	}
	if c.RefreshToken == "" {
		missing = append(missing, "FREEAGENT_REFRESH_TOKEN")
	}
	if c.ClientID == "" {
		missing = append(missing, "FREEAGENT_CLIENT_ID")
	}
	if c.ClientSecret == "" {
		missing = append(missing, "FREEAGENT_CLIENT_SECRET")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing environment variables: %s", strings.Join(missing, ", "))
	}
	return c, nil
}

func (c *Client) refreshAccessToken() error {
	// Create the refresh URL
	form := fmt.Sprintf(
		"grant_type=refresh_token&refresh_token=%s",
		c.RefreshToken,
	)

	// Setup the http request
	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/token_endpoint",
		strings.NewReader(form),
	)
	if err != nil {
		return err
	}

	// Encode the creds and set headers
	creds := base64.StdEncoding.EncodeToString([]byte(c.ClientID + ":" + c.ClientSecret))

	req.Header.Set("Authorization", "Basic "+creds)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Do the http request
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read token response body: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("token refresh failed %d: %s", res.StatusCode, string(body))
	}

	var data struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	// Make sure that the return body is valid
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	c.AccessToken = data.AccessToken

	// Check if the refresh token matches the existing one and if not inform the user to update.
	if data.RefreshToken != "" && data.RefreshToken != c.RefreshToken {
		fmt.Fprintln(os.Stderr, "New refresh token issued. Update FREEAGENT_REFRESH_TOKEN:")
		fmt.Fprintln(os.Stderr, data.RefreshToken)
	}

	return nil
}

func (c *Client) apiRequest(method, path string, body any, out any) error {
	var reqBody io.Reader

	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(b)
	}

	url := path
	// Check that the URL is https
	if !strings.HasPrefix(path, "http") {
		url = baseURL + path
	}

	// Prepair http request
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return err
	}

	// Set Headers
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Go-Freeagent-SDK/1.0")

	// Do the http request
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Handle Auth code expiration
	if res.StatusCode == http.StatusUnauthorized {
		fmt.Fprintln(os.Stderr, "Access token expired, refreshing...")

		if err := c.refreshAccessToken(); err != nil {
			return err
		}

		return c.apiRequest(method, path, body, out)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	// check the request worked
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("api error %d %s %s: %s", res.StatusCode, method, url, string(resBody))
	}

	if len(resBody) == 0 || out == nil {
		return nil
	}

	return json.Unmarshal(resBody, out)
}

// type Config struct {
// 	UserID string
// 	ProjectID string
// 	TaskID string
// }

// func cmdListTimesheets(c *Client, start_date string, end_date string, cfg)

func cmdListProjects(c *Client, view string) error {
	endpoint := "/projects"
	if view != "" {
		q := url.Values{}
		q.Set("view", view)
		endpoint = endpoint + "?" + q.Encode()
	}
	var result struct {
		Projects []struct {
			URL    string `json:"url"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"projects"`
	}

	if err := c.apiRequest(http.MethodGet, endpoint, nil, &result); err != nil {
		return err
	}

	for _, p := range result.Projects {
		id := path.Base(p.URL)
		fmt.Printf("%-6s  %-30s  %-20s\n", id, p.Name, p.Status)
	}

	return nil
}

func main() {
	client, err := newFaClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error", err)
		os.Exit(1)
	}

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage: freeagent")
		os.Exit(0)
	}

	command := args[0]
	switch command {
	case "projects":
		if err := cmdListProjects(client, ""); err != nil {
			fmt.Fprintln(os.Stderr, "error", err)
			os.Exit(1)
		}
	}
}
