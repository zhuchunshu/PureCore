package oauth

import (
	"fmt"

	"purecore/core"
)

// BaseProvider implements common methods shared by most OAuth providers.
// Embed it in your provider struct and override methods as needed.
//
// By default, it assumes the provider is a standard OAuth2 provider:
//   - IsOAuth2() returns true
//   - ConfigFields() returns the standard OAuth2 fields
//   - HandleCallback() returns an error (not applicable)
//
// Non-OAuth2 providers (e.g., Telegram) should override IsOAuth2() and HandleCallback().
type BaseProvider struct{}

// GetConfig reads a configuration value for the current provider from the global
// admin option store. The concrete provider must implement ConfigKeyPrefix().
// For a provider to use this method, it should call the package-level helper
// GetProviderOption instead, which takes the Provider interface.
func (p BaseProvider) GetConfig(key string) string {
	// Cannot determine prefix without concrete type; return empty.
	// Providers should call GetProviderOption(provider, key) instead.
	return ""
}

// FullConfigKey appends the provider prefix to the given key, producing the
// full key used in the WebOption table (e.g. "oauth_github_client_id").
func FullConfigKey(provider Provider, key string) string {
	return provider.ConfigKeyPrefix() + "_" + key
}

// GetProviderOption retrieves a configuration value from the admin option store
// for the given provider and key suffix. This is the recommended way for providers
// to read their own settings.
func GetProviderOption(provider Provider, key string) string {
	fullKey := FullConfigKey(provider, key)
	return core.AdminOption(fullKey, "")
}

// ---------- Default implementations for the Provider interface ----------

// ConfigFields returns the standard set of OAuth2 configuration fields.
// Override this method for providers that need different fields (e.g., Telegram).
func (p BaseProvider) ConfigFields() []ConfigField {
	return BaseOAuth2ConfigFields()
}

// IsOAuth2 returns true; override for non-OAuth2 providers like Telegram.
func (p BaseProvider) IsOAuth2() bool {
	return true
}

// GetDocURL returns an empty string by default; override in concrete providers.
func (p BaseProvider) GetDocURL() string {
	return ""
}

// GetApplyURL returns an empty string by default; override in concrete providers.
func (p BaseProvider) GetApplyURL() string {
	return ""
}

// HandleCallback is not supported for standard OAuth2 providers.
// Non-OAuth2 providers must override this to handle their custom callback logic.
func (p BaseProvider) HandleCallback(params map[string]string) (*UserInfo, error) {
	return nil, fmt.Errorf("HandleCallback not supported for this provider")
}
