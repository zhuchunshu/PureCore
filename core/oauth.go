package core

import (
	"github.com/markbates/goth"
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
	if AdminOption("oauth_github_enabled") == "1" {
		registerProvider("github", appURL)
	}
	if AdminOption("oauth_google_enabled") == "1" {
		registerProvider("google", appURL)
	}
	if AdminOption("oauth_discord_enabled") == "1" {
		registerProvider("discord", appURL)
	}

	oauthInitialized = true
}

func registerProvider(provider, appURL string) {
	clientID := AdminOption("oauth_" + provider + "_client_id")
	clientSecret := AdminOption("oauth_" + provider + "_client_secret")
	if clientID == "" || clientSecret == "" {
		return
	}

	callbackURL := appURL + "/api/v1/oauth/" + provider + "/callback"

	switch provider {
	case "github":
		goth.UseProviders(
			github.New(clientID, clientSecret, callbackURL, "read:user", "user:email"),
		)
	case "google":
		goth.UseProviders(
			google.New(clientID, clientSecret, callbackURL, "email", "profile"),
		)
	case "discord":
		goth.UseProviders(
			discord.New(clientID, clientSecret, callbackURL, discord.ScopeIdentify, discord.ScopeEmail),
		)
	}
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
	}

	var result []map[string]string
	for _, p := range providers {
		if AdminOption("oauth_"+p.key+"_enabled") == "1" &&
			AdminOption("oauth_"+p.key+"_client_id") != "" &&
			AdminOption("oauth_"+p.key+"_client_secret") != "" {
			result = append(result, map[string]string{
				"key":  p.key,
				"name": p.name,
			})
		}
	}
	return result
}
