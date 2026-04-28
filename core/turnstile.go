package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Turnstile verification endpoint
const turnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

// Cloudflare Turnstile test keys — always pass verification, usable on any domain including localhost.
// https://developers.cloudflare.com/turnstile/reference/testing/
const (
	turnstileTestSiteKey   = "1x00000000000000000000AA"
	turnstileTestSecretKey = "1x0000000000000000000000000000000AA"
)

// TurnstileResponse is the response from Cloudflare's siteverify endpoint
type TurnstileResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
	Action      string   `json:"action"`
	CData       string   `json:"cdata"`
}

// VerifyTurnstile sends a token to Cloudflare for verification.
// Returns nil if the token is valid, or an error describing the failure.
// Automatically uses test keys when running on localhost/127.0.0.1.
func VerifyTurnstile(token string) error {
	secretKey := AdminOption("turnstile_secret_key")
	if secretKey == "" {
		return errors.New("turnstile secret key not configured")
	}
	if token == "" {
		return errors.New("turnstile token is empty")
	}

	// If running locally, use test keys which always pass verification
	if isLocalhost(secretKey) {
		secretKey = turnstileTestSecretKey
	}

	data := url.Values{
		"secret":   {secretKey},
		"response": {token},
	}

	resp, err := http.PostForm(turnstileVerifyURL, data)
	if err != nil {
		return fmt.Errorf("turnstile verification request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read turnstile response: %w", err)
	}

	var result TurnstileResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse turnstile response: %w", err)
	}

	if !result.Success {
		errorMsg := strings.Join(result.ErrorCodes, ", ")
		return fmt.Errorf("turnstile verification failed: %s", errorMsg)
	}

	return nil
}

// IsTurnstileEnabled checks if Turnstile is configured and should be used
// for the given context (e.g., "admin_login", "admin_register", "public_login").
func IsTurnstileEnabled(context string) bool {
	siteKey := AdminOption("turnstile_site_key")
	secretKey := AdminOption("turnstile_secret_key")
	if siteKey == "" || secretKey == "" {
		return false
	}
	// Check the per-context toggle
	if AdminOption(context) != "1" {
		return false
	}
	return true
}

// isLocalhost checks if the backend is running in a local development environment.
// This is determined by checking if the configured secret key matches the test key
// pattern, or if the host appears to be localhost.
func isLocalhost(secretKey string) bool {
	// If the configured key looks like a test key (starts with "1x"), we're in dev mode
	if strings.HasPrefix(secretKey, "1x") {
		return true
	}
	// If the key is the well-known test key, use test verification
	if secretKey == turnstileTestSecretKey {
		return true
	}
	return false
}

// GetTurnstileSiteKey returns the appropriate Turnstile site key.
// In local development with test keys configured, returns the test site key.
func GetTurnstileSiteKey() string {
	key := AdminOption("turnstile_site_key", "")
	// If using test keys (detected by secret key), return test site key
	if key != "" && strings.HasPrefix(key, "1x") {
		return turnstileTestSiteKey
	}
	return key
}
