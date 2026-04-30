# Authentication System

PureCore implements a dual JWT authentication system with access tokens and refresh tokens for both regular users and administrators.

## Overview

| User Type | Access Token (JWT) | Refresh Token | Password Hashing |
|-----------|-------------------|---------------|------------------|
| Regular User | 15 minutes | 7 days, stored in DB | bcrypt |
| Admin | 15 minutes | 7 days, stored in DB | bcrypt |

Both user types use separate JWT secrets and token generation functions, ensuring complete isolation between user and admin authentication.

## User Authentication

### Registration

Regular users register via `POST /api/v1/auth/register`. The system:

1. Validates the request body (`name`, `email`, `password`)
2. Optionally verifies a Turnstile token (if `turnstile_public_login` option is enabled)
3. Checks for duplicate email addresses
4. Hashes the password with bcrypt
5. Creates the user record
6. Generates an access token (JWT, 15 min) and refresh token (random string, 7 days)
7. Saves the refresh token to the `users` table

**User model fields:**

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| name | string | Display name |
| email | string | Email (unique) |
| password | string | Bcrypt hash |
| avatar | string | Avatar URL (nullable) |
| bio | string | User bio (nullable) |
| status | int | Account status |
| email_verified_at | time | Email verification timestamp (nullable) |
| last_login_at | time | Last login timestamp (nullable) |
| refresh_token | string | Current refresh token |
| created_at | time | Creation timestamp |
| updated_at | time | Update timestamp |
| deleted_at | time | Soft delete timestamp |

### Login

Users login via `POST /api/v1/auth/login`. The system:

1. Validates credentials
2. Optionally verifies Turnstile
3. Looks up the user by email
4. Verifies the password hash
5. Generates new access and refresh tokens
6. Updates `refresh_token` and `last_login_at` in the database

### Token Refresh

Users refresh their access token via `POST /api/v1/auth/refresh`. The system:

1. Validates the provided refresh token
2. Looks up the user by refresh_token
3. Generates a new access token
4. Rotates the refresh token (generates new, replaces old in DB) for security

### Profile

Authenticated users can access their profile via `GET /api/v1/auth/profile`. The endpoint returns all profile fields including avatar, bio, status, and timestamps.

## Admin Authentication

### Admin Registration

Admins register via `POST /{admin_prefix}/auth/register`. The system:

1. Validates the request body (`username`, `password`, `name`)
2. Optionally verifies Turnstile (if `turnstile_admin_register` option is enabled)
3. Assigns the `"admin"` role (always)
4. Creates the admin record
5. Returns access and refresh tokens

**Admin model fields:**

| Field | Type | Description |
|-------|------|-------------|
| id | uint | Primary key |
| username | string | Login username (unique) |
| name | string | Display name |
| password | string | Bcrypt hash |
| role | string | Role: `"admin"` or `"super_admin"` |
| token_version | int | Incremented on password change to invalidate tokens |
| refresh_token | string | Current refresh token |
| created_at | time | Creation timestamp |
| updated_at | time | Update timestamp |

### Admin Login

Admins login via `POST /{admin_prefix}/auth/login`. The flow is identical to user login but uses the admin-specific token generation (`GenerateAdminToken`).

The admin JWT includes additional claims:

```json
{
  "id": 1,
  "username": "admin",
  "role": "admin",
  "token_version": 0,
  "exp": 1234567890
}
```

The `token_version` claim is checked by the `AdminAuth` middleware. If it doesn't match the current value in the database, the token is rejected — this enables forced logout on password change.

### Password Change & Token Invalidation

Admins can change their password via `POST /{admin_prefix}/auth/change-password`. This endpoint:

1. Validates current and new passwords
2. Verifies the current password matches
3. Hashes the new password
4. Increments `token_version` (invalidates all existing JWTs)
5. Clears `refresh_token` (invalidates all refresh tokens)

After this operation, the admin must re-login to obtain new tokens.

### Admin Refresh Token

Admins refresh their access token via `POST /{admin_prefix}/auth/refresh`. The flow is identical to user token refresh but uses admin-specific token generation.

## Middleware

### Auth (User)

Located at `app/Http/Middleware/Auth.go`. This middleware:

1. Extracts the Bearer token from the `Authorization` header
2. Verifies the JWT using `GenerateUserToken`
3. Stores the user's claims in `c.Locals("user")` as `map[string]string{"id": "..."}`
4. Returns 401 if the token is missing, invalid, or expired

### AdminAuth

Located at `app/Http/Middleware/AdminAuth.go`. This middleware:

1. Extracts the Bearer token from the `Authorization` header
2. Verifies the JWT using `GenerateAdminToken`
3. Checks `token_version` against the database — rejects if mismatched
4. Stores admin claims in `c.Locals("user")`
5. Returns 401 if any check fails

### Named Middleware Registration

Both middleware are registered as named middleware in the `RouteServiceProvider`:

```go
router.RegisterNamedMiddlewares(map[string]core.MiddlewareFunc{
    "auth":       middleware.Auth(),
    "admin_auth": middleware.AdminAuth(),
})
```

## JWT Token Generation

### User Tokens

```go
// In app/Http/Middleware/Auth.go
func GenerateUserToken(userID uint, userName string) (string, error)
```

Claims: `id`, `name`, `exp` (15 minutes from now).
Uses `JWT_SECRET` from environment variables.

### Admin Tokens

```go
// In app/Http/Middleware/AdminAuth.go
func GenerateAdminToken(adminID uint, username string, tokenVersion int) (string, error)
```

Claims: `id`, `username`, `role`, `token_version`, `exp` (15 minutes from now).
Uses `JWT_SECRET` from environment variables.

### Refresh Tokens

```go
func GenerateRefreshToken() (string, error)
```

Generates a random 64-character hex string. Valid for 7 days (enforced by the client, not cryptographically).

## Client-Side Authentication

### User Authentication (useUserAuth.js)

The `web/src/composables/useUserAuth.js` composable handles:

- Login, register, and logout flows
- Token storage in localStorage
- Automatic token refresh
- Reactive user state

### Admin Authentication (useAuth.js)

The `web/src/composables/useAuth.js` composable handles:

- Admin login, register, and logout flows
- Token storage in localStorage
- Automatic token refresh before expiry
- Reactive admin user state

## Security Considerations

1. **Password Hashing**: All passwords are hashed with bcrypt before storage
2. **Token Rotation**: Refresh tokens are rotated on each use — old tokens are immediately invalidated
3. **Token Versioning**: Admin JWTs include a `token_version` claim — changing password invalidates all tokens across all devices
4. **Short-Lived Access Tokens**: 15-minute expiry minimizes the window for token theft
5. **Unique Secrets**: Use a strong, unique `JWT_SECRET` in production (never use the default)
6. **Turnstile Protection**: Optionally require Cloudflare Turnstile verification for login/register endpoints
