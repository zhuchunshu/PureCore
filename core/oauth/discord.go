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

// DiscordProvider implements the OAuth2 flow for Discord.
// Docs: https://discord.com/developers/docs/topics/oauth2
// App creation: https://discord.com/developers/applications
type DiscordProvider struct {
	BaseProvider
}

func init() {
	Register(&DiscordProvider{})
}

func (p *DiscordProvider) Name() string            { return "discord" }
func (p *DiscordProvider) DisplayName() string     { return "Discord" }
func (p *DiscordProvider) ConfigKeyPrefix() string { return "oauth_discord" }

func (p *DiscordProvider) GetDocURL() string {
	return "https://discord.com/developers/docs/topics/oauth2"
}

func (p *DiscordProvider) GetApplyURL() string {
	return "https://discord.com/developers/applications"
}

// GetAuthURL builds the Discord OAuth2 authorize URL.
func (p *DiscordProvider) GetAuthURL(state string) string {
	clientID := GetProviderOption(p, "client_id")
	redirectURL := GetProviderOption(p, "redirect_url")
	u, _ := url.Parse("https://discord.com/api/oauth2/authorize")
	q := u.Query()
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURL)
	q.Set("response_type", "code")
	q.Set("scope", "identify email")
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String()
}

// Exchange exchanges the authorization code for an access token and fetches user info.
func (p *DiscordProvider) Exchange(ctx context.Context, code string) (*UserInfo, error) {
	clientID := GetProviderOption(p, "client_id")
	clientSecret := GetProviderOption(p, "client_secret")
	redirectURL := GetProviderOption(p, "redirect_url")

	// Step 1: Exchange code for access token
	token, err := p.exchangeCode(ctx, clientID, clientSecret, redirectURL, code)
	if err != nil {
		return nil, fmt.Errorf("discord: failed to exchange code: %w", err)
	}

	// Step 2: Fetch user info
	user, err := p.fetchUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("discord: failed to fetch user: %w", err)
	}

	return &UserInfo{
		ProviderID: user.ID,
		Email:      user.Email,
		Name:       user.GlobalName,
		AvatarURL:  p.buildAvatarURL(user),
		Raw:        user,
	}, nil
}

// ---------- Discord API helpers ----------

type discordTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

type discordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	GlobalName    string `json:"global_name"`
	Avatar        string `json:"avatar"`
	Email         string `json:"email"`
	Verified      bool   `json:"verified"`
	Discriminator string `json:"discriminator"`
}

func (p *DiscordProvider) exchangeCode(ctx context.Context, clientID, clientSecret, redirectURL, code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://discord.com/api/oauth2/token", strings.NewReader(data.Encode()))
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

	var tr discordTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}
	if tr.Error != "" {
		return "", fmt.Errorf("%s: %s", tr.Error, tr.ErrorDesc)
	}
	return tr.AccessToken, nil
}

func (p *DiscordProvider) fetchUser(ctx context.Context, token string) (*discordUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://discord.com/api/users/@me", nil)
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

	var user discordUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode user: %w", err)
	}
	return &user, nil
}

func (p *DiscordProvider) buildAvatarURL(user *discordUser) string {
	if user.Avatar == "" {
		return ""
	}
	format := "png"
	if len(user.Avatar) >= 2 && user.Avatar[:2] == "a_" {
		format = "gif"
	}
	return "https://cdn.discordapp.com/avatars/" + user.ID + "/" + user.Avatar + "." + format
}
