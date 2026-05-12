package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Login runs the OAuth 2.0 authorization-code flow against FreeAgent:
//   - Starts a local HTTP server on the given port
//   - Opens the user's default browser to the approval URL
//   - Captures the callback, exchanges the code for tokens
//   - Saves the tokens to the cache file
func Login(clientID, clientSecret string, port int) error {
	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("FREEAGENT_CLIENT_ID and FREEAGENT_CLIENT_SECRET must be set")
	}

	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)
	state, err := randomState()
	if err != nil {
		return err
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			errCh <- fmt.Errorf("state mismatch — possible CSRF attempt")
			return
		}
		if errParam := q.Get("error"); errParam != "" {
			desc := q.Get("error_description")
			http.Error(w, "authorization denied: "+errParam, http.StatusBadRequest)
			errCh <- fmt.Errorf("authorization failed: %s (%s)", errParam, desc)
			return
		}
		code := q.Get("code")
		if code == "" {
			http.Error(w, "no code in callback", http.StatusBadRequest)
			errCh <- fmt.Errorf("no code in callback")
			return
		}
		fmt.Fprintln(w, "Authentication successful! You can close this window.")
		codeCh <- code
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("callback server: %w", err)
		}
	}()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	authURL := buildAuthURL(clientID, redirectURI, state)
	fmt.Fprintln(os.Stderr, "Opening browser for authentication...")
	fmt.Fprintln(os.Stderr, "If it doesn't open, visit:")
	fmt.Fprintln(os.Stderr, authURL)

	if err := openBrowser(authURL); err != nil {
		fmt.Fprintf(os.Stderr, "(could not open browser automatically: %v)\n", err)
	}

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		return err
	case <-time.After(5 * time.Minute):
		return fmt.Errorf("timeout waiting for authentication")
	}

	tokens, err := exchangeCodeForTokens(code, clientID, clientSecret, redirectURI)
	if err != nil {
		return err
	}

	if err := saveTokenCache(tokens); err != nil {
		return fmt.Errorf("save tokens: %w", err)
	}
	return nil
}

func buildAuthURL(clientID, redirectURI, state string) string {
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("state", state)
	return BaseURL + "/approve_app?" + q.Encode()
}

func exchangeCodeForTokens(code, clientID, clientSecret, redirectURI string) (tokenCache, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest(http.MethodPost, BaseURL+"/token_endpoint", strings.NewReader(form.Encode()))
	if err != nil {
		return tokenCache{}, err
	}
	creds := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+creds)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return tokenCache{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return tokenCache{}, fmt.Errorf("read token response: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return tokenCache{}, fmt.Errorf("token exchange failed %d: %s", res.StatusCode, body)
	}

	var data struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return tokenCache{}, err
	}
	return tokenCache{
		AccessToken:  data.AccessToken,
		RefreshToken: data.RefreshToken,
	}, nil
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	return cmd.Start()
}
