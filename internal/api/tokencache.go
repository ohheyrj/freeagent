package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
)

type tokenCache struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func tokenCachePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	u, err := url.Parse(BaseURL)
	if err != nil {
		return "", err
	}
	name := fmt.Sprintf("tokens-%s.json", u.Host)
	return filepath.Join(cacheDir, "freeagent", name), nil
}

func loadTokenCache() (tokenCache, error) {
	path, err := tokenCachePath()
	if err != nil {
		return tokenCache{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return tokenCache{}, nil
		}
		return tokenCache{}, err
	}
	var tc tokenCache
	return tc, json.Unmarshal(data, &tc)
}

func saveTokenCache(tc tokenCache) error {
	path, err := tokenCachePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.Marshal(tc)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
