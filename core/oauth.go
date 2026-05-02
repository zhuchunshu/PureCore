package core

import (
	"time"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/apple"
	"github.com/markbates/goth/providers/discord"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

var oauthInitialized bool

// InitOAuth registers all enabled OAuth providers with Goth.
// Must be called after database is available (i.e., after Boot).
// appURL is the base URL of the application (e.g., "http://localhost:9002").
func InitOAuth(appURL string) {
	if oauthInitialized {
		return
	}

	// Register providers based on WebOption flags
	if shouldRegisterProvider("github") {
		registerProvider("github", appURL)
	}
	if shouldRegisterProvider("google") {
		registerProvider("google", appURL)
	}
	if shouldRegisterProvider("discord") {
		registerProvider("discord", appURL)
	}
	if shouldRegisterProvider("apple") {
		registerProvider("apple", appURL)
	}
	if shouldRegisterProvider("telegram") {
		registerTelegramProvider(appURL)
	}

	oauthInitialized = true
}

// shouldRegisterProvider returns true if the provider is enabled (login or register allowed)
// and has the necessary credentials configured.
func shouldRegisterProvider(provider string) bool {
	if !ProviderActive(provider) {
		return false
	}
	switch provider {
	case "telegram":
		return AdminOption("oauth_telegram_bot_token") != ""
	case "apple":
		return AdminOption("oauth_apple_client_id") != "" &&
			AdminOption("oauth_apple_team_id") != "" &&
			AdminOption("oauth_apple_key_id") != "" &&
			AdminOption("oauth_apple_private_key") != ""
	default:
		return AdminOption("oauth_"+provider+"_client_id") != "" &&
			AdminOption("oauth_"+provider+"_client_secret") != ""
	}
}

// ProviderActive checks if the provider has any auth flow enabled (login or register).
// It also considers the legacy "enabled" flag for backward compatibility.
func ProviderActive(provider string) bool {
	login := AdminOption("oauth_" + provider + "_allow_login")
	register := AdminOption("oauth_" + provider + "_allow_register")
	legacy := AdminOption("oauth_" + provider + "_enabled")

	// If new toggles are explicitly set, use them
	if login != "" || register != "" {
		return login == "1" || register == "1"
	}
	// Fallback to legacy enabled flag
	return legacy == "1"
}

// OAuthAllowLogin returns true if login is allowed for the given provider.
func OAuthAllowLogin(provider string) bool {
	if !ProviderActive(provider) {
		return false
	}
	login := AdminOption("oauth_" + provider + "_allow_login")
	if login != "" {
		return login == "1"
	}
	// Fallback: if legacy enabled is set, allow login
	return AdminOption("oauth_"+provider+"_enabled") == "1"
}

// OAuthAllowRegister returns true if registration of new users is allowed for the given provider.
func OAuthAllowRegister(provider string) bool {
	if !ProviderActive(provider) {
		return false
	}
	register := AdminOption("oauth_" + provider + "_allow_register")
	if register != "" {
		return register == "1"
	}
	// Fallback: if legacy enabled is set, allow registration
	return AdminOption("oauth_"+provider+"_enabled") == "1"
}

func registerProvider(provider, appURL string) {
	defaultCallback := appURL + "/api/v1/oauth/" + provider + "/callback"
	callbackURL := AdminOption("oauth_" + provider + "_callback")
	if callbackURL == "" {
		callbackURL = defaultCallback
	}

	switch provider {
	case "github":
		clientID := AdminOption("oauth_github_client_id")
		clientSecret := AdminOption("oauth_github_client_secret")
		goth.UseProviders(
			github.New(clientID, clientSecret, callbackURL, "read:user", "user:email"),
		)
	case "google":
		clientID := AdminOption("oauth_google_client_id")
		clientSecret := AdminOption("oauth_google_client_secret")
		goth.UseProviders(
			google.New(clientID, clientSecret, callbackURL, "email", "profile"),
		)
	case "discord":
		clientID := AdminOption("oauth_discord_client_id")
		clientSecret := AdminOption("oauth_discord_client_secret")
		goth.UseProviders(
			discord.New(clientID, clientSecret, callbackURL, discord.ScopeIdentify, discord.ScopeEmail),
		)
	case "apple":
		clientID := AdminOption("oauth_apple_client_id")
		teamID := AdminOption("oauth_apple_team_id")
		keyID := AdminOption("oauth_apple_key_id")
		privateKey := AdminOption("oauth_apple_private_key")

		// Generate the client secret JWT using Apple's MakeSecret
		now := time.Now()
		secret, err := apple.MakeSecret(apple.SecretParams{
			PKCS8PrivateKey: privateKey,
			TeamId:          teamID,
			KeyId:           keyID,
			ClientId:        clientID,
			Iat:             int(now.Unix()),
			Exp:             int(now.Add(5 * time.Minute).Unix()),
		})
		if err != nil || secret == nil {
			return
		}
		goth.UseProviders(
			apple.New(clientID, *secret, callbackURL, nil, apple.ScopeName, apple.ScopeEmail),
		)
	}
}

// registerTelegramProvider registers the custom Telegram provider.
// Telegram uses a widget-based auth flow and redirects to a callback URL
// with a query parameter containing the authorization data.
func registerTelegramProvider(appURL string) {
	botToken := AdminOption("oauth_telegram_bot_token")
	defaultCallback := appURL + "/api/v1/oauth/telegram/callback"
	callbackURL := AdminOption("oauth_telegram_callback")
	if callbackURL == "" {
		callbackURL = defaultCallback
	}
	tgProv := NewTelegramProvider(botToken, callbackURL)
	goth.UseProviders(tgProv)
}

// EnabledOAuthProviders returns the list of providers that are configured and enabled.
func EnabledOAuthProviders() []map[string]string {
	providers := []struct {
		key  string
		name string
	}{
		{"github", "GitHub"},
		{"google", "Google"},
		{"discord", "Discord"},
		{"apple", "Apple"},
		{"telegram", "Telegram"},
	}

	var result []map[string]string
	for _, p := range providers {
		if !ProviderActive(p.key) {
			continue
		}
		// Check credentials
		hasCreds := false
		switch p.key {
		case "telegram":
			hasCreds = AdminOption("oauth_telegram_bot_token") != ""
		case "apple":
			hasCreds = AdminOption("oauth_apple_client_id") != "" &&
				AdminOption("oauth_apple_team_id") != "" &&
				AdminOption("oauth_apple_key_id") != "" &&
				AdminOption("oauth_apple_private_key") != ""
		default:
			hasCreds = AdminOption("oauth_"+p.key+"_client_id") != "" &&
				AdminOption("oauth_"+p.key+"_client_secret") != ""
		}
		if !hasCreds {
			continue
		}
		result = append(result, map[string]string{
			"key":            p.key,
			"name":           p.name,
			"allow_login":    AdminOption("oauth_" + p.key + "_allow_login"),
			"allow_register": AdminOption("oauth_" + p.key + "_allow_register"),
			"callback_url":   AdminOption("oauth_" + p.key + "_callback"),
		})
	}
	return result
}
