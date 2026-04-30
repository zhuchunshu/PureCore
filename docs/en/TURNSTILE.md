# Turnstile Integration

PureCore supports [Cloudflare Turnstile](https://developers.cloudflare.com/turnstile/) — a privacy-friendly CAPTCHA alternative — to protect public-facing authentication endpoints from bots and automated abuse.

## Overview

Turnstile verification can be enabled on a per-context basis for these scenarios:

| Context Key | Endpoints Protected |
|-------------|---------------------|
| `turnstile_public_login` | `/api/v1/auth/login`, `/api/v1/auth/register` |
| `turnstile_admin_login` | `/{admin_prefix}/auth/login` |
| `turnstile_admin_register` | `/{admin_prefix}/auth/register` |

Each context is independently toggleable via admin options.

## How It Works

### Flow

1. The frontend renders a Turnstile widget (using `TurnstileWidget.vue`) when the corresponding option is enabled
2. The user completes the Turnstile challenge
3. On form submission, the Turnstile token is sent alongside the other form data (e.g., `turnstile_token` field)
4. The backend calls Cloudflare's siteverify endpoint to validate the token
5. If verification succeeds, the request proceeds; otherwise, a 422 error is returned

### Backend Implementation

The Turnstile logic is centralized in `core/turnstile.go`:

```go
// Verify a Turnstile token
func VerifyTurnstile(token string) error

// Check if Turnstile is enabled for a given context
func IsTurnstileEnabled(context string) bool

// Get the appropriate site key (returns test key in dev mode)
func GetTurnstileSiteKey() string
```

**Controller integration:**

```go
func (uc *UserAuthController) Login(req *core.Request, res *core.Response) error {
    var body UserLoginRequest
    if err := req.Validate(&body); err != nil {
        return res.Error("Invalid credentials", 422)
    }

    // Verify Turnstile if enabled for public login
    if core.IsTurnstileEnabled("turnstile_public_login") {
        if err := core.VerifyTurnstile(body.TurnstileToken); err != nil {
            return res.Error("Captcha verification failed: "+err.Error(), 422)
        }
    }
    // ... proceed with authentication
}
```

### Frontend Implementation

The `web/src/components/TurnstileWidget.vue` component:

- Loads the Cloudflare Turnstile script dynamically
- Renders the widget when the site key is available
- Emits a `verified` event with the token on successful challenge completion
- Supports resetting the widget for re-verification (e.g., after form submission failure)

**Usage in a login/register form:**

```vue
<script setup>
import TurnstileWidget from '@/components/TurnstileWidget.vue'

const turnstileToken = ref('')

function handleSubmit() {
  // Send turnstileToken.value along with form data
}
</script>

<template>
  <TurnstileWidget
    v-if="turnstileEnabled"
    @verified="token => turnstileToken = token"
  />
</template>
```

## Configuration

### Admin Options

Turnstile is configured via the admin options system (`web_options` table). An administrator can set these options through the admin settings panel or via the API:

| Option Key | Description | Example Value |
|------------|-------------|---------------|
| `turnstile_site_key` | Cloudflare Turnstile site key | `1x00000000000000000000AA` |
| `turnstile_secret_key` | Cloudflare Turnstile secret key | `1x0000000000000000000000000000000AA` |
| `turnstile_public_login` | Enable Turnstile for public user login/register | `"1"` (enabled) or `""` (disabled) |
| `turnstile_admin_login` | Enable Turnstile for admin login | `"1"` or `""` |
| `turnstile_admin_register` | Enable Turnstile for admin registration | `"1"` or `""` |

**Setting options via API:**

```bash
curl -X POST \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"turnstile_site_key":"1x00000000000000000000AA","turnstile_secret_key":"1x0000000000000000000000000000000AA","turnstile_public_login":"1"}' \
  http://localhost:9002/control-panel/options
```

### Checking Turnstile Status (Public API)

The frontend can check whether Turnstile is enabled and get the site key via the public options endpoint:

```bash
curl http://localhost:9002/control-panel/options
```

Response includes `turnstile_site_key` if configured. The frontend checks for the presence of this key and the corresponding context toggle to decide whether to render the widget.

## Development & Testing

### Test Keys

Cloudflare provides **test keys** that always pass verification and work on any domain, including `localhost`:

| Key Type | Value |
|----------|-------|
| Site Key | `1x00000000000000000000AA` |
| Secret Key | `1x0000000000000000000000000000000AA` |

The backend automatically detects test keys (any key starting with `1x`) and uses the official test secret key when verifying tokens. This means:

- Set the test keys as admin options during development
- The widget will always pass verification
- No requests are sent to Cloudflare's servers with test keys

**Example development setup:**

```
turnstile_site_key=1x00000000000000000000AA
turnstile_secret_key=1x0000000000000000000000000000000AA
turnstile_public_login=1
```

### Disabling Turnstile

To disable Turnstile for a specific context, either:

1. Set the context option to an empty string or remove it
2. Remove the site key and secret key (disables all Turnstile functionality)

## Security Notes

1. **Server-side verification is mandatory**: Never trust the client to verify Turnstile — always use `core.VerifyTurnstile()` on the backend
2. **Secret key protection**: The secret key is stored in the `web_options` database table (accessible only to authenticated admins). Never expose it in client-side code.
3. **Test key detection**: The backend only uses test secret keys when the configured key starts with `1x`. In production, always use real Cloudflare keys.
4. **Token is single-use**: Each Turnstile token can only be verified once. Forms must generate a new token on each submission attempt.
