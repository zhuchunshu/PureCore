package oauth

import (
	"context"
)

// UserInfo represents the normalized user profile returned by an OAuth provider.
type UserInfo struct {
	ProviderID string `json:"provider_id"` // Unique ID from the provider (e.g., GitHub user ID)
	Email      string `json:"email"`
	Name       string `json:"name"`
	AvatarURL  string `json:"avatar_url"`
	Raw        any    `json:"raw"` // Raw provider-specific data, for future extensibility
}

// ConfigField defines a single configuration field for an OAuth provider's admin settings tab.
// Each provider can declare its own set of fields — some need client_id + client_secret,
// others may need bot_token, team_id, private_key, etc.
type ConfigField struct {
	Key         string `json:"key"`         // Config key suffix (e.g., "client_id", "bot_token")
	Label       string `json:"label"`       // Human-readable label for the admin form
	Type        string `json:"type"`        // "text", "password", "toggle"
	InputType   string `json:"input_type"`  // Optional: "textarea" for multi-line, empty for default input
	Required    bool   `json:"required"`    // Whether this field must be filled to enable the provider
	Placeholder string `json:"placeholder"` // Placeholder text for the input
	Help        string `json:"help"`        // Help text / usage hint shown below the input
}

// Provider defines the interface that every OAuth provider must implement.
//
// To add a new provider (e.g., Google, Facebook):
//  1. Create a struct in core/oauth/ that embeds BaseProvider and implements this interface.
//  2. Register it in an init() function via oauth.Register().
//  3. No changes to routes, controllers, or frontend are needed — they auto-detect registered providers.
type Provider interface {
	// Name returns a unique provider identifier (e.g., "github", "google").
	// This value is used in routes, database records, and admin settings keys.
	Name() string

	// DisplayName returns a human-readable name for the UI (e.g., "GitHub", "Google").
	DisplayName() string

	// GetAuthURL generates the authorization URL that the user's browser should be redirected to.
	// The state parameter must be used for CSRF protection and eventual redirect back to the frontend.
	GetAuthURL(state string) string

	// Exchange handles the callback from the provider: it exchanges the authorization code
	// for an access token, fetches the user profile, and returns a normalized UserInfo.
	Exchange(ctx context.Context, code string) (*UserInfo, error)

	// ConfigKeyPrefix returns the prefix for keys in the WebOption table / admin settings
	// that belong to this provider (e.g., "oauth_github").
	ConfigKeyPrefix() string

	// GetConfig retrieves a configuration value for this provider.
	// The implementation is already provided by BaseProvider; you can override if needed.
	GetConfig(key string) string

	// ConfigFields returns the list of configuration fields this provider needs in the admin panel.
	// Each field defines its key, type, label, etc. The frontend renders the form dynamically.
	ConfigFields() []ConfigField

	// IsOAuth2 returns true if this provider follows the standard OAuth2 authorization-code flow.
	// Non-OAuth2 providers (e.g., Telegram) return false and may need custom callback handling.
	IsOAuth2() bool

	// GetDocURL returns the URL to the provider's developer documentation / app creation page.
	// Shown in the admin OAuth settings tab for this provider.
	GetDocURL() string

	// GetApplyURL returns the URL where admins can create/register an OAuth application for this provider.
	// Shown in the admin OAuth settings tab.
	GetApplyURL() string

	// HandleCallback processes the callback data for non-OAuth2 providers (e.g., Telegram).
	// The params map contains the query parameters from the callback request.
	// OAuth2 providers should return an error indicating this method is not supported.
	HandleCallback(params map[string]string) (*UserInfo, error)
}

// ---------- Default config fields ----------

// BaseOAuth2ConfigFields returns the standard set of configuration fields for OAuth2 providers.
func BaseOAuth2ConfigFields() []ConfigField {
	return []ConfigField{
		{Key: "enabled", Label: "admin.oauth_enabled", Type: "toggle", Required: false},
		{Key: "login_enabled", Label: "admin.oauth_login_enabled", Type: "toggle", Required: false},
		{Key: "register_enabled", Label: "admin.oauth_register_enabled", Type: "toggle", Required: false},
		{Key: "client_id", Label: "admin.oauth_client_id", Type: "text", Required: true},
		{Key: "client_secret", Label: "admin.oauth_client_secret", Type: "password", Required: true},
		{Key: "redirect_url", Label: "admin.oauth_redirect_url", Type: "text", Required: true},
	}
}

// RedirectURLPath returns the callback path for the given provider.
// Example: "/oauth/github/callback"
func RedirectURLPath(providerName string) string {
	return "/oauth/" + providerName + "/callback"
}
