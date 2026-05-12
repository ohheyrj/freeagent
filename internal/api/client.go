package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const productionURL = "https://api.freeagent.com/v2"

var BaseURL = func() string {
	if u := os.Getenv("FREEAGENT_BASE_URL"); u != "" {
		return u
	}
	return productionURL
}()

type Client struct {
	HTTPClient   *http.Client
	AccessToken  string
	RefreshToken string
	ClientID     string
	ClientSecret string
}

func NewClient() (*Client, error) {
	if BaseURL != productionURL {
		fmt.Fprintf(os.Stderr, "Using FreeAgent API: %s\n", BaseURL)
	}

	c := &Client{
		HTTPClient:   http.DefaultClient,
		AccessToken:  os.Getenv("FREEAGENT_ACCESS_TOKEN"),
		RefreshToken: os.Getenv("FREEAGENT_REFRESH_TOKEN"),
		ClientID:     os.Getenv("FREEAGENT_CLIENT_ID"),
		ClientSecret: os.Getenv("FREEAGENT_CLIENT_SECRET"),
	}

	if cached, err := loadTokenCache(); err == nil {
		if cached.AccessToken != "" {
			c.AccessToken = cached.AccessToken
		}
		if cached.RefreshToken != "" {
			c.RefreshToken = cached.RefreshToken
		}
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
		BaseURL+"/token_endpoint",
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
	if data.RefreshToken != "" {
		c.RefreshToken = data.RefreshToken
	}

	if err := saveTokenCache(tokenCache{
		AccessToken:  c.AccessToken,
		RefreshToken: c.RefreshToken,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "warning: could not write token cache:", err)
	}

	return nil
}

func (c *Client) DoRequest(method, path string, body any) (*http.Response, []byte, error) {
	var reqBody io.Reader

	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		reqBody = bytes.NewReader(b)
	}

	url := path
	// Check that the URL is https
	if !strings.HasPrefix(path, "http") {
		url = BaseURL + path
	}

	// Prepair http request
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, nil, err
	}

	// Set Headers
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Go-Freeagent-SDK/1.0")

	// Do the http request
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	// Handle Auth code expiration
	if res.StatusCode == http.StatusUnauthorized {
		fmt.Fprintln(os.Stderr, "Access token expired, refreshing...")

		if err := c.refreshAccessToken(); err != nil {
			return nil, nil, err
		}

		return c.DoRequest(method, path, body)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response body: %w", err)
	}

	// check the request worked
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, nil, fmt.Errorf("api error %d %s %s: %s", res.StatusCode, method, url, string(resBody))
	}

	return res, resBody, nil
}

func (c *Client) APIRequest(method, path string, body any, out any) error {
	_, resBody, err := c.DoRequest(method, path, body)
	if err != nil {
		return err
	}
	if out == nil || len(resBody) == 0 {
		return nil
	}
	return json.Unmarshal(resBody, out)
}
