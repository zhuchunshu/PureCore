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

// Provider defines the interface that every OAuth provider must implement.
//
// To add a new provider (e.g., Google, Facebook):
//  1. Create a struct in core/oauth/ that embeds BaseProvider and implements this interface.
//  2. Register it in core/oauth/registry.go via the init() or exported function.
//  3. Add its configuration namespace in core/oauth/config.go (optional, if needs custom keys).
//  4. Create the admin settings tab on the frontend – the tab is driven by the registered provider list.
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
}
