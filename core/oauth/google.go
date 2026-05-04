package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GoogleProvider implements the OAuth2 flow for Google.
// Docs: https://developers.google.com/identity/protocols/oauth2
// App creation: https://console.cloud.google.com/apis/credentials
type GoogleProvider struct {
	BaseProvider
}

func init() {
	Register(&GoogleProvider{})
}

func (p *GoogleProvider) Name() string            { return "google" }
func (p *GoogleProvider) DisplayName() string     { return "Google" }
func (p *GoogleProvider) ConfigKeyPrefix() string { return "oauth_google" }

func (p *GoogleProvider) GetDocURL() string {
	return "https://developers.google.com/identity/protocols/oauth2"
}

func (p *GoogleProvider) GetApplyURL() string {
	return "https://console.cloud.google.com/apis/credentials"
}

// GetAuthURL builds the Google OAuth2 authorize URL.
func (p *GoogleProvider) GetAuthURL(state string) string {
	clientID := GetProviderOption(p, "client_id")
	redirectURL := GetProviderOption(p, "redirect_url")
	u, _ := url.Parse("https://accounts.google.com/o/oauth2/v2/auth")
	q := u.Query()
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURL)
	q.Set("response_type", "code")
	q.Set("scope", "openid profile email")
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String()
}

// Exchange exchanges the authorization code for an access token and fetches user info.
func (p *GoogleProvider) Exchange(ctx context.Context, code string) (*UserInfo, error) {
	clientID := GetProviderOption(p, "client_id")
	clientSecret := GetProviderOption(p, "client_secret")
	redirectURL := GetProviderOption(p, "redirect_url")

	// Step 1: Exchange code for access token
	token, err := p.exchangeCode(ctx, clientID, clientSecret, redirectURL, code)
	if err != nil {
		return nil, fmt.Errorf("google: failed to exchange code: %w", err)
	}

	// Step 2: Fetch user info
	user, err := p.fetchUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("google: failed to fetch user: %w", err)
	}

	return &UserInfo{
		ProviderID: user.Sub,
		Email:      user.Email,
		Name:       user.Name,
		AvatarURL:  user.Picture,
		Raw:        user,
	}, nil
}

// ---------- Google API helpers ----------

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

type googleUserInfo struct {
	Sub     string `json:"sub"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

func (p *GoogleProvider) exchangeCode(ctx context.Context, clientID, clientSecret, redirectURL, code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURL)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tr googleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}
	if tr.Error != "" {
		return "", fmt.Errorf("%s: %s", tr.Error, tr.ErrorDesc)
	}
	return tr.AccessToken, nil
}

func (p *GoogleProvider) fetchUserInfo(ctx context.Context, token string) (*googleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode user info: %w", err)
	}
	return &user, nil
}
