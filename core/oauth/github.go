package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GitHubProvider implements the OAuth2 flow for GitHub.
// Docs: https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps
// App creation: https://github.com/settings/developers
type GitHubProvider struct {
	BaseProvider
}

func init() {
	Register(&GitHubProvider{})
}

func (p *GitHubProvider) Name() string            { return "github" }
func (p *GitHubProvider) DisplayName() string     { return "GitHub" }
func (p *GitHubProvider) ConfigKeyPrefix() string { return "oauth_github" }

func (p *GitHubProvider) GetDocURL() string {
	return "https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps"
}

func (p *GitHubProvider) GetApplyURL() string {
	return "https://github.com/settings/developers"
}

// GetAuthURL builds the GitHub OAuth authorize URL.
func (p *GitHubProvider) GetAuthURL(state string) string {
	clientID := GetProviderOption(p, "client_id")
	redirectURL := GetProviderOption(p, "redirect_url")
	u, _ := url.Parse("https://github.com/login/oauth/authorize")
	q := u.Query()
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURL)
	q.Set("state", state)
	q.Set("scope", "read:user user:email")
	u.RawQuery = q.Encode()
	return u.String()
}

// Exchange exchanges the OAuth code for an access token and fetches the user info.
func (p *GitHubProvider) Exchange(ctx context.Context, code string) (*UserInfo, error) {
	clientID := GetProviderOption(p, "client_id")
	clientSecret := GetProviderOption(p, "client_secret")

	// Step 1: Exchange code for access token
	token, err := p.exchangeCode(ctx, clientID, clientSecret, code)
	if err != nil {
		return nil, fmt.Errorf("github: failed to exchange code: %w", err)
	}

	// Step 2: Fetch user profile
	user, err := p.fetchUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("github: failed to fetch user: %w", err)
	}

	// Step 3: Fetch user emails (primary email may not be in profile)
	email := user.Email
	if email == "" {
		email, _ = p.fetchPrimaryEmail(ctx, token)
	}

	return &UserInfo{
		ProviderID: fmt.Sprintf("%d", user.ID),
		Email:      email,
		Name:       user.Login,
		AvatarURL:  user.AvatarURL,
		Raw:        user,
	}, nil
}

// ---------- GitHub API helpers ----------

type githubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

type githubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (p *GitHubProvider) exchangeCode(ctx context.Context, clientID, clientSecret, code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tr githubTokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", fmt.Errorf("parse error: %w, body: %s", err, string(body))
	}
	if tr.Error != "" {
		return "", fmt.Errorf("%s: %s", tr.Error, tr.ErrorDesc)
	}
	return tr.AccessToken, nil
}

func (p *GitHubProvider) fetchUser(ctx context.Context, token string) (*githubUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user githubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode user: %w", err)
	}
	return &user, nil
}

func (p *GitHubProvider) fetchPrimaryEmail(ctx context.Context, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []githubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	// Fallback to first verified email
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}
	return "", fmt.Errorf("no verified email found")
}
