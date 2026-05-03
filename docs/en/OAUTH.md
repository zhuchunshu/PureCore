# OAuth / Third-Party Login

PureCore provides a production-ready, extensible OAuth framework for adding third-party login providers (GitHub, Google, etc.). The architecture follows the **Strategy Pattern** — each provider implements a common `Provider` interface, making it trivial to add new platforms without modifying existing code.

## Architecture Overview

| Layer | File(s) | Role |
|-------|---------|------|
| **Interface** | `core/oauth/base.go` | Defines the `Provider` contract (authorize, exchange token, fetch user info) |
| **Registry** | `core/oauth/registry.go` | Global provider registry; providers self-register via `init()` |
| **Base Provider** | `core/oauth/provider.go` | Partial implementation that new providers can embed |
| **State Management** | `core/oauth/state.go` | CSRF-safe state tokens (HMAC-SHA256 nonces, 10-minute expiry) |
| **Database Model** | `app/Models/OAuthAccount.go` | Maps OAuth provider + external user ID → PureCore user ID |
| **Migration** | `database/migrations/2026_05_03_000000_create_oauth_accounts_table.go` | Schema for `oauth_accounts` table |
| **Controller** | `app/Http/Controllers/OAuthController.go` | Redirect (start OAuth flow), Callback (handle return), Bind/Unbind endpoints |
| **Routes** | `routes/oauth.go` | Registers all OAuth endpoints with their middlewares |
| **Frontend Composable** | `web/src/composables/useOAuth.js` | Fetches available providers and builds redirect URLs |
| **Frontend Component** | `web/src/components/auth/OAuthButton.vue` | Reusable button for any registered provider |
| **Callback Page** | `web/src/pages/auth/OAuthCallback.vue` | Handles OAuth callback in the browser |
| **Admin Settings** | `web/src/pages/admin/AdminOAuthSettings.vue` | Admin panel to enable/disable providers and configure credentials |

## Provider Interface

Every OAuth provider must implement the `Provider` interface defined in `core/oauth/base.go`:

```go
type Provider interface {
    // Name returns the provider identifier (e.g., "github", "google")
    Name() string

    // DisplayName returns the human-readable name shown on buttons and settings
    DisplayName() string

    // GetAuthURL generates the authorization URL with state and redirect
    GetAuthURL(state string, redirectURL string) string

    // ExchangeCode exchanges an authorization code for an access token
    ExchangeCode(code string, redirectURL string) (*Token, error)

    // GetUserInfo uses the access token to fetch the user's profile
    GetUserInfo(token *Token) (*UserInfo, error)
}

type Token struct {
    AccessToken  string
    TokenType    string
    Refreshed     string // optional
    ExpiresIn    int
}

type UserInfo struct {
    ProviderUserID string // unique ID on the provider's side
    Email          string
    Name           string
    AvatarURL      string
    RawData        map[string]interface{} // raw response for custom use
}
```

The `Provider` interface is the only contract a new provider needs to fulfil. A partial base implementation is provided in `core/oauth/provider.go` that can be embedded to reduce boilerplate for common OAuth 2.0 flows.

## Registry Pattern

The global registry (`core/oauth/registry.go`) allows providers to register themselves without touching any existing code:

```go
var registry = make(map[string]Provider)

func Register(p Provider) {
    registry[p.Name()] = p
}

func Get(name string) (Provider, bool) {
    p, ok := registry[name]
    return p, ok
}

func All() map[string]Provider {
    return registry
}
```

Providers typically call `Register()` inside an `init()` function, so they are automatically available when the binary starts.

## Adding a New Provider (e.g., GitHub)

To add GitHub login, you create a single new file:

### Step 1: Create the provider implementation

Create `core/oauth/github.go`:

```go
package oauth

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
)

type GitHubProvider struct {
    *BaseProvider
}

func init() {
    Register(&GitHubProvider{})
}

func (p *GitHubProvider) Name() string { return "github" }
func (p *GitHubProvider) DisplayName() string { return "GitHub" }

func (p *GitHubProvider) GetAuthURL(state string, redirectURL string) string {
    u := "https://github.com/login/oauth/authorize"
    return u + "?client_id=" + p.getSetting("client_id") +
        "&redirect_uri=" + url.QueryEscape(redirectURL) +
        "&scope=user:email&state=" + state
}

func (p *GitHubProvider) ExchangeCode(code string, redirectURL string) (*Token, error) {
    // POST https://github.com/login/oauth/access_token
    // with client_id, client_secret, code, redirect_uri
    // Parse response to obtain access_token
}

func (p *GitHubProvider) GetUserInfo(token *Token) (*UserInfo, error) {
    // GET https://api.github.com/user with Authorization: Bearer token
    // Parse JSON to fill UserInfo
}
```

### Step 2: Register the settings keys

The admin panel expects each provider's configuration to be stored as web options. By convention, the keys follow the pattern:

- `oauth_{provider}_enabled` — "1" or "0"
- `oauth_{provider}_login_enabled` — "1" or "0"
- `oauth_{provider}_register_enabled` — "1" or "0"
- `oauth_{provider}_client_id`
- `oauth_{provider}_client_secret`
- `oauth_{provider}_redirect_url`

The `BaseProvider` helper (`core/oauth/provider.go`) provides a `getSetting(key)` method that reads these values from the option store.

### Step 3: Run migration (if any additional tables needed)

Most providers work fine with the existing `oauth_accounts` table. If a provider requires extra data, create a new migration.

### That's it!

After adding the provider file, it is automatically available — no route changes, no controller modifications, no frontend updates. The admin settings page and login/register buttons will detect and display it automatically.

## How the OAuth Flow Works

1. **User clicks "Login with X"** on the login or register page.
   - `OAuthButton.vue` uses `useOAuth.js` to build a URL like `/api/v1/auth/oauth/github/redirect?redirect=/dashboard`.
   - The `redirect` query param is where the user will be sent after successful login.

2. **Redirect endpoint** (`GET /api/v1/auth/oauth/{provider}/redirect`):
   - Looks up the provider from the registry.
   - Generates a signed state token (CSRF protection).
   - Stores the `redirect` target in the state token.
   - Redirects the browser to the provider's authorization page.

3. **User authorizes** on the provider's website.

4. **Provider redirects back** to the configured redirect URL (`GET /api/v1/oauth/callback?code=...&state=...`).
   - `OAuthCallback.vue` extracts `code` and `state` params.
   - Calls the backend at `POST /api/v1/auth/oauth/callback` with the `code` and `state`.

5. **Backend Callback endpoint** (`POST /api/v1/auth/oauth/callback`):
   - Verifies the state token (CSRF check, expiry check).
   - Exchanges the authorization code for an access token via `provider.ExchangeCode()`.
   - Fetches user info via `provider.GetUserInfo()`.
   - Checks if an `oauth_account` record exists for this provider + provider user ID:
     - **If found**: Issues a JWT + refresh token for the linked user and returns `{ action: "logged_in", token, refresh_token }`.
     - **If not found**: Returns `{ action: "bind", oauth_id, provider, user_info: { email, name, avatar_url } }`.

6. **Frontend handles the "bind" case** in `OAuthCallback.vue`:
   - Shows two options:
     1. **Register with provider info**: Pre-fills the registration form with the provider's email and name. After successful registration, the backend automatically links the OAuth account.
     2. **Login and bind**: The user logs in with existing credentials. After login, the OAuth account is linked.

7. **Bind endpoint** (`POST /api/v1/auth/oauth/bind`):
   - Requires the user to be authenticated (JWT required).
   - Links the OAuth account to the currently logged-in user.

## Admin Settings

Administrators can configure each OAuth provider via the **OAuth Settings** page (`/{admin_prefix}/oauth`). The page provides:

- **Tabbed interface**: Each registered provider appears as a tab.
- **Feature toggles**: Enable/disable the provider, and separately control whether it appears on the login page and/or register page.
- **Credentials**: Client ID, Client Secret, and Redirect URL fields.
- All settings are stored as web options (via `app/Models/WebOption`) and automatically respected by the authentication system.

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/v1/auth/oauth/{provider}/redirect` | None | Redirect user to provider's authorization page |
| `GET` | `/api/v1/oauth/callback` | None | Callback URL (page) after provider authorizes |
| `POST` | `/api/v1/auth/oauth/callback` | None | Exchange code and process login/bind |
| `POST` | `/api/v1/auth/oauth/bind` | User JWT | Link OAuth account to authenticated user |
| `DELETE` | `/api/v1/auth/oauth/unbind/{provider}` | User JWT | Unlink an OAuth account |
| `GET` | `/api/v1/auth/oauth/accounts` | User JWT | List linked OAuth accounts for the user |
| `GET` | `/api/v1/auth/oauth/providers` | None | List available OAuth providers (for frontend buttons) |
| `GET` | `/api/v1/{admin_prefix}/oauth/settings` | Admin JWT | Get OAuth provider settings for admin |
| `POST` | `/api/v1/{admin_prefix}/oauth/settings` | Admin JWT | Update OAuth provider settings |

## Security Considerations

1. **CSRF Protection**: State tokens are HMAC-SHA256 signed and expire after 10 minutes. The backend validates state on every callback.
2. **Settings Isolation**: OAuth credentials (Client ID, Secret) are stored in the database and only accessible to administrators.
3. **User Privacy**: The `OAuthAccount` model stores only the provider name and external user ID — raw OAuth user data is not persisted.
4. **Token Security**: OAuth access tokens are never stored server-side. They are used only to fetch user info during the callback and then discarded.
5. **Admin-Only Configuration**: Only authenticated administrators can view or modify OAuth provider settings.
