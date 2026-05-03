package oauth

// BaseProvider implements common shared by most OAuth providers.
// Embed it in your provider struct and override methods as needed.
//
// Example usage for a GitHub provider:
//
//	type GitHubProvider struct {
//	    BaseProvider
//	}
//
//	func (p *GitHubProvider) Name() string       { return "github" }
//	func (p *GitHubProvider) DisplayName() string { return "GitHub" }
//	func (p *GitHubProvider) GetAuthURL(state string) string {
//	    // Build the GitHub authorize URL using client_id from config
//	    return "https://github.com/login/oauth/authorize?..."
//	}
//	func (p *GitHubProvider) Exchange(ctx context.Context, code string) (*UserInfo, error) {
//	    // Exchange code for token, fetch user info, return normalized UserInfo
//	}
type BaseProvider struct{}

// GetConfig reads a configuration value for the current provider from the global
// admin option store. It uses the provider's ConfigKeyPrefix() to namespace the key.
//
// For example, a GitHub provider with ConfigKeyPrefix() == "oauth_github" calling
// p.GetConfig("client_id") will look up the option "oauth_github_client_id".
func (p BaseProvider) GetConfig(key string) string {
	return "" // Default: no config, override in concrete provider
}

// FullConfigKey appends the provider prefix to the given key, producing the
// full key used in the WebOption table (e.g. "oauth_github_client_id").
func FullConfigKey(provider Provider, key string) string {
	return provider.ConfigKeyPrefix() + "_" + key
}
