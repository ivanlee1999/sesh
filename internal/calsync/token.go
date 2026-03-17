package calsync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"

	"github.com/ivanlee1999/sesh/internal/config"
)

// TokenPath returns the path to the token file for the given provider.
func TokenPath(provider string) string {
	return filepath.Join(config.DataDir(), provider+"_token.json")
}

// SaveToken writes an OAuth2 token to disk as JSON.
func SaveToken(provider string, token *oauth2.Token) error {
	path := TokenPath(provider)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create token dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open token file: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

// LoadToken reads an OAuth2 token from disk.
func LoadToken(provider string) (*oauth2.Token, error) {
	path := TokenPath(provider)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read token file: %w", err)
	}
	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	return &token, nil
}

// HasToken checks if a token file exists for the given provider.
func HasToken(provider string) bool {
	_, err := os.Stat(TokenPath(provider))
	return err == nil
}
