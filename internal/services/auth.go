// ==========================
// internal/services/auth.go
// ==========================
package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"awesomeProject/internal/config"
)

type TokenCache struct {
	token  string
	expiry time.Time
	mu     sync.RWMutex
}

var tokenCache = &TokenCache{}

func (tc *TokenCache) Set(token string, expiresIn int) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.token = token
	tc.expiry = time.Now().Add(time.Duration(expiresIn-60) * time.Second)
}

func (tc *TokenCache) Get() (string, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	if tc.token != "" && time.Now().Before(tc.expiry) {
		return tc.token, true
	}
	return "", false
}

func (tc *TokenCache) Clear() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.token = ""
	tc.expiry = time.Time{}
}

type AuthService struct {
	config *config.Config
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{config: cfg}
}

func (s *AuthService) GetAccessToken(forceRefresh bool) (string, error) {
	if !forceRefresh {
		if token, ok := tokenCache.Get(); ok {
			return token, nil
		}
	}

	authString := fmt.Sprintf("%s:%s", s.config.ConsumerKey, s.config.ConsumerSecret)
	encoded := base64.StdEncoding.EncodeToString([]byte(authString))

	req, err := http.NewRequest("GET", s.config.OAuthURL(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+encoded)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(s.config.APITimeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to generate access token: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   string `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.AccessToken == "" {
		return "", fmt.Errorf("access token not found in response")
	}

	expiresIn := 3600
	if result.ExpiresIn != "" {
		fmt.Sscanf(result.ExpiresIn, "%d", &expiresIn)
	}

	tokenCache.Set(result.AccessToken, expiresIn)
	return result.AccessToken, nil
}
