package oauth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TelegramProvider implements Telegram Login Widget (non-OAuth2).
// Docs: https://core.telegram.org/widgets/login
// Bot creation: https://t.me/BotFather
type TelegramProvider struct {
	BaseProvider
}

func init() {
	Register(&TelegramProvider{})
}

func (p *TelegramProvider) Name() string            { return "telegram" }
func (p *TelegramProvider) DisplayName() string     { return "Telegram" }
func (p *TelegramProvider) ConfigKeyPrefix() string { return "oauth_telegram" }

func (p *TelegramProvider) GetDocURL() string {
	return "https://core.telegram.org/widgets/login"
}

func (p *TelegramProvider) GetApplyURL() string {
	return "https://t.me/BotFather"
}

// IsOAuth2 returns false — Telegram does not use the OAuth2 authorization code flow.
func (p *TelegramProvider) IsOAuth2() bool {
	return false
}

// ConfigFields returns the fields needed for Telegram (no client_id/client_secret, uses bot_token).
func (p *TelegramProvider) ConfigFields() []ConfigField {
	return []ConfigField{
		{Key: "enabled", Label: "admin.oauth_enabled", Type: "toggle", Required: false},
		{Key: "login_enabled", Label: "admin.oauth_login_enabled", Type: "toggle", Required: false},
		{Key: "register_enabled", Label: "admin.oauth_register_enabled", Type: "toggle", Required: false},
		{Key: "bot_token", Label: "admin.oauth_telegram_bot_token", Type: "password", Required: true, Placeholder: "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11", Help: "admin.oauth_telegram_bot_token_help"},
		{Key: "redirect_url", Label: "admin.oauth_redirect_url", Type: "text", Required: true, Help: "admin.oauth_redirect_url_help"},
	}
}

// GetAuthURL returns an empty string; Telegram uses a JavaScript widget, not a redirect.
// The frontend should handle Telegram login via the widget.
func (p *TelegramProvider) GetAuthURL(state string) string {
	// Telegram login is handled by the frontend widget; the auth URL is not used.
	// However, we still generate a URL for compatibility: the frontend can use it as
	// a fallback or for direct linking.
	botToken := GetProviderOption(p, "bot_token")
	redirectURL := GetProviderOption(p, "redirect_url")
	if botToken == "" {
		return ""
	}
	// The bot username is derived from the token; frontend can parse it.
	// For the state parameter, we'll embed it in the redirect URL query.
	// Actually Telegram login widget handles this; we just provide the redirect URL.
	// Return the redirect URL with state appended.
	if state != "" {
		if strings.Contains(redirectURL, "?") {
			redirectURL += "&state=" + state
		} else {
			redirectURL += "?state=" + state
		}
	}
	return redirectURL
}

// Exchange is not used for Telegram; the provider uses HandleCallback instead.
func (p *TelegramProvider) Exchange(ctx context.Context, code string) (*UserInfo, error) {
	return nil, fmt.Errorf("Telegram does not use OAuth2 code exchange; use HandleCallback instead")
}

// HandleCallback verifies the Telegram login data and returns user info.
// The params map should contain the query parameters from the callback,
// including: id, first_name, last_name, username, photo_url, auth_date, hash.
// The hash is verified using the bot token.
func (p *TelegramProvider) HandleCallback(params map[string]string) (*UserInfo, error) {
	botToken := GetProviderOption(p, "bot_token")
	if botToken == "" {
		return nil, fmt.Errorf("telegram: bot_token not configured")
	}

	// Verify the hash
	if err := p.verifyTelegramAuth(params, botToken); err != nil {
		return nil, fmt.Errorf("telegram: verification failed: %w", err)
	}

	// Build UserInfo
	id := params["id"]
	firstName := params["first_name"]
	lastName := params["last_name"]
	username := params["username"]
	photoURL := params["photo_url"]

	name := firstName
	if lastName != "" {
		name += " " + lastName
	}
	if name == "" {
		name = username
	}

	// Email is not provided by Telegram; we'll leave it empty
	email := ""
	// Some bots may send email if the user has shared it; not standard

	rawData, _ := json.Marshal(params)

	return &UserInfo{
		ProviderID: id,
		Email:      email,
		Name:       name,
		AvatarURL:  photoURL,
		Raw:        rawData,
	}, nil
}

// verifyTelegramAuth verifies the hash from Telegram login data.
// It follows the algorithm described at https://core.telegram.org/widgets/login#checking-authorization
func (p *TelegramProvider) verifyTelegramAuth(params map[string]string, botToken string) error {
	// Extract hash
	receivedHash := params["hash"]
	if receivedHash == "" {
		return fmt.Errorf("missing hash parameter")
	}

	// Build data-check-string
	// 1. Sort keys alphabetically
	keys := make([]string, 0, len(params)-1)
	for k := range params {
		if k == "hash" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 2. Concatenate key=value pairs with \n
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	dataCheckString := strings.Join(parts, "\n")

	// 3. Compute HMAC-SHA256 of the bot token to get the secret key
	secretKey := sha256Hmac([]byte("WebAppData"), []byte(botToken))

	// 4. Compute HMAC-SHA256 of the data-check-string with the secret key
	computedHash := sha256Hmac(secretKey, []byte(dataCheckString))

	// 5. Compare hex of computed hash with received hash
	computedHex := hex.EncodeToString(computedHash)
	if !hmac.Equal([]byte(computedHex), []byte(receivedHash)) {
		return fmt.Errorf("hash mismatch")
	}

	// Also check auth_date is not too old (within 1 day)
	if authDateStr := params["auth_date"]; authDateStr != "" {
		authDate, err := strconv.ParseInt(authDateStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid auth_date")
		}
		now := time.Now().Unix()
		if now-authDate > 86400 { // 1 day in seconds
			return fmt.Errorf("auth_date is too old (more than 24 hours)")
		}
	}

	return nil
}

// sha256Hmac computes HMAC-SHA256 and returns the raw bytes.
func sha256Hmac(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
