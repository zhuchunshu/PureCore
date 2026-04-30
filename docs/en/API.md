# PureCore API Documentation

## General Information

- Base URL: `/api/v1`
- Admin Base URL: `/{admin_prefix}` (default: `/control-panel`, configurable via `ADMIN_ROUTE_PREFIX`)
- Default Port: `9002`
- Authentication: `Bearer Token` (JWT)
- Response Format: JSON

### Unified Response Structure

All endpoints return a unified JSON format:

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": {}
}
```

| Field | Type | Description |
|------|------|------|
| code | int | Status code, 0 for success, non-zero for error |
| message | string | Message, supports Chinese and English |
| data | object/array | Response data, may be null |

### Paginated Response

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": [],
  "total": 100,
  "page": 1,
  "per_page": 15
}
```

| Field | Type | Description |
|------|------|------|
| total | int | Total number of records |
| page | int | Current page number |
| per_page | int | Items per page |

### Error Status Codes

| HTTP Status | code | Description |
|-------------|------|------|
| 200 | 0 | Success |
| 400 | 400 | Bad request |
| 401 | 401 | Unauthorized |
| 404 | 404 | Not found |
| 409 | 409 | Conflict (e.g., duplicate email) |
| 422 | 422 | Validation failed |
| 500 | 500 | Internal server error |

## Public Endpoints

### Health Check

```
GET /api/v1/ping
```

**Example Request**

```bash
curl http://localhost:9002/api/v1/ping
```

**Example Response**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": "pong"
}
```

### Project Information

```
GET /api/v1/system/info
```

Returns project metadata from `purecore.json`, including version, release type, author, license, and dependencies.

**Example Request**

```bash
curl http://localhost:9002/api/v1/system/info
```

**Example Response**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": {
    "name": "PureCore",
    "version": "1.0.0",
    "release_type": "alpha",
    "author": {
      "name": "zhuchunshu",
      "email": "",
      "url": "https://github.com/zhuchunshu"
    },
    "repository": {
      "type": "git",
      "url": "https://github.com/zhuchunshu/PureCore.git"
    },
    "license": "MIT",
    "keywords": ["go", "gofiber", "vue", "laravel-style", "framework"],
    "go_version": "1.21",
    "dependencies": {
      "backend": {
        "framework": "GoFiber v3",
        "database": "PostgreSQL",
        "orm": "GORM",
        "validation": "go-playground/validator"
      },
      "frontend": {
        "framework": "Vue 3",
        "build_tool": "Vite",
        "css": "Tailwind CSS + DaisyUI",
        "package_manager": "Bun"
      }
    },
    "description": {
      "en": "A full-stack Go web development framework wrapping GoFiber v3 in a Laravel-like development style.",
      "zh": "基于 Go 语言的全栈 Web 开发框架，将 GoFiber v3 封装成类似 Laravel 的开发风格。"
    }
  }
}
```

| Field | Type | Description |
|------|------|------|
| name | string | Project name |
| version | string | Semantic version (e.g. 1.0.0) |
| release_type | string | Release stage: `alpha`, `beta`, `rc`, or `stable` |
| author | object | Author info (name, email, url) |
| license | string | License identifier |
| keywords | array | Project tags |
| description | object | Localized descriptions (keyed by locale) |

### Documentation

```
GET /api/v1/docs?file=en/README.md
```

Returns the content of a documentation file from the `docs/` directory.

**Query Parameters**

| Parameter | Type | Required | Description |
|------|------|------|------|
| file | string | Yes | Path to the doc file relative to `docs/` (e.g. `en/README.md`) |

**Example Request**

```bash
curl "http://localhost:9002/api/v1/docs?file=en/README.md"
```

**Example Response**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": {
    "filename": "en/README.md",
    "content": "# PureCore Framework\n\n..."
  }
}
```

### Documentation List

```
GET /api/v1/docs/list
```

Returns a list of all available documentation files.

**Example Request**

```bash
curl http://localhost:9002/api/v1/docs/list
```

**Example Response**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": ["en/README.md", "en/API.md", "zh/README.md", "zh/API.md"]
}
```

## Public Options

```
GET /{admin_prefix}/options
```

Returns all public web options (site name, description, keywords, Turnstile site key, etc.). This endpoint does **not** require authentication.

**Example Request**

```bash
curl http://localhost:9002/control-panel/options
```

**Example Response**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": {
    "site_name": "PureCore",
    "site_description": "A full-stack Go framework",
    "site_keywords": "go, vue, framework",
    "turnstile_site_key": "1x00000000000000000000AA"
  }
}
```

## User Authentication

### Register

```
POST /api/v1/auth/register
```

Creates a new regular user account. Returns JWT access token and refresh token on success.

**Request Body**

| Parameter | Type | Required | Description |
|------|------|------|------|
| name | string | Yes | Username, minimum 2 characters |
| email | string | Yes | Email, must be valid format and unique |
| password | string | Yes | Password, minimum 6 characters |
| turnstile_token | string | Conditional | Turnstile token (required if Turnstile is enabled for public login) |

**Example Request**

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","password":"password123"}' \
  http://localhost:9002/api/v1/auth/register
```

**Example Response (Success)**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "abc123def456...",
    "name": "Alice",
    "email": "alice@example.com"
  }
}
```

**Example Response (Email Exists)**

```json
{
  "code": 409,
  "message": "Email already exists"
}
```

### Login

```
POST /api/v1/auth/login
```

Authenticates a regular user and returns JWT tokens. Updates `last_login_at` timestamp.

**Request Body**

| Parameter | Type | Required | Description |
|------|------|------|------|
| email | string | Yes | Registered email |
| password | string | Yes | Password, minimum 6 characters |
| turnstile_token | string | Conditional | Turnstile token (required if Turnstile is enabled for public login) |

**Example Request**

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"password123"}' \
  http://localhost:9002/api/v1/auth/login
```

**Example Response (Success)**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "abc123def456...",
    "name": "Alice",
    "email": "alice@example.com"
  }
}
```

**Example Response (Invalid Credentials)**

```json
{
  "code": 401,
  "message": "Invalid credentials"
}
```

### Refresh Token

```
POST /api/v1/auth/refresh
```

Generates a new access token using a valid refresh token. The old refresh token is rotated (replaced with a new one) for security.

**Request Body**

| Parameter | Type | Required | Description |
|------|------|------|------|
| refresh_token | string | Yes | Valid refresh token from login/register |

**Example Request**

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"abc123def456..."}' \
  http://localhost:9002/api/v1/auth/refresh
```

**Example Response (Success)**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "new-refresh-token..."
  }
}
```

### Get Profile

```
GET /api/v1/auth/profile
```

Returns the authenticated user's profile. Requires `Authorization: Bearer <token>` header.

**Example Request**

```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  http://localhost:9002/api/v1/auth/profile
```

**Example Response**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": {
    "id": 1,
    "name": "Alice",
    "email": "alice@example.com",
    "avatar": null,
    "bio": null,
    "status": 1,
    "email_verified_at": null,
    "last_login_at": "2026-04-30T10:30:00Z",
    "created_at": "2026-04-28T08:00:00Z",
    "updated_at": "2026-04-30T10:30:00Z"
  }
}
```

## Admin Authentication

> All admin endpoints use a configurable prefix (default: `/control-panel`, set via `ADMIN_ROUTE_PREFIX`).

### Check Admin Exists

```
GET /{admin_prefix}/auth/check
```

Returns whether any admin users exist. Used by the frontend to determine whether to show the registration or login page.

**Example Request**

```bash
curl http://localhost:9002/control-panel/auth/check
```

**Example Response**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": {
    "exists": true,
    "count": 1
  }
}
```

### Register Admin

```
POST /{admin_prefix}/auth/register
```

Creates a new admin user. Always assigns the `"admin"` role. First admin creation does not require prior authentication.

**Request Body**

| Parameter | Type | Required | Description |
|------|------|------|------|
| username | string | Yes | Admin username, minimum 3 characters |
| password | string | Yes | Password, minimum 6 characters |
| name | string | Yes | Display name |
| turnstile_token | string | Conditional | Turnstile token (required if Turnstile is enabled for admin register) |

**Example Request**

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123","name":"Administrator"}' \
  http://localhost:9002/control-panel/auth/register
```

**Example Response (Success)**

```json
{
  "code": 0,
  "message": "Admin registered successfully",
  "data": {
    "message": "Admin registered successfully",
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "abc123def456...",
    "username": "admin",
    "name": "Administrator",
    "role": "admin"
  }
}
```

### Admin Login

```
POST /{admin_prefix}/auth/login
```

Authenticates an admin user and returns JWT tokens. The access token includes `token_version` to support forced logout (changing password increments `token_version`, invalidating all existing tokens).

**Request Body**

| Parameter | Type | Required | Description |
|------|------|------|------|
| username | string | Yes | Admin username, minimum 3 characters |
| password | string | Yes | Password, minimum 6 characters |
| turnstile_token | string | Conditional | Turnstile token (required if Turnstile is enabled for admin login) |

**Example Request**

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  http://localhost:9002/control-panel/auth/login
```

**Example Response (Success)**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "abc123def456...",
    "username": "admin",
    "name": "Administrator",
    "role": "admin"
  }
}
```

### Admin Refresh Token

```
POST /{admin_prefix}/auth/refresh
```

Generates a new access token using a valid refresh token. The old refresh token is rotated.

**Request Body**

| Parameter | Type | Required | Description |
|------|------|------|------|
| refresh_token | string | Yes | Valid refresh token from login/register |

**Example Request**

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"abc123def456..."}' \
  http://localhost:9002/control-panel/auth/refresh
```

**Example Response (Success)**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "new-refresh-token..."
  }
}
```

### Get Admin Profile

```
GET /{admin_prefix}/auth/profile
```

Returns the authenticated admin's profile. Requires admin authentication (`Authorization: Bearer <admin_token>`).

**Example Request**

```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  http://localhost:9002/control-panel/auth/profile
```

**Example Response**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": {
    "id": 1,
    "username": "admin",
    "name": "Administrator",
    "role": "admin"
  }
}
```

### Change Admin Password

```
POST /{admin_prefix}/auth/change-password
```

Changes the current admin's password. **Increments `token_version`**, which invalidates all existing tokens (including the current one). The client must re-login after this call.

**Request Body**

| Parameter | Type | Required | Description |
|------|------|------|------|
| current_password | string | Yes | Current password, minimum 6 characters |
| new_password | string | Yes | New password, minimum 6 characters |

**Example Request**

```bash
curl -X POST \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{"current_password":"admin123","new_password":"newpass456"}' \
  http://localhost:9002/control-panel/auth/change-password
```

**Example Response (Success)**

```json
{
  "code": 0,
  "message": "Password changed successfully",
  "data": {
    "message": "Password changed successfully"
  }
}
```

### Admin Options (Set)

```
POST /{admin_prefix}/options
```

Sets web options (key-value pairs). Requires admin authentication. Accepts a JSON object where keys are option names and values are option values.

**Request Body**

| Parameter | Type | Required | Description |
|------|------|------|------|
| (dynamic keys) | string | No | Any option key-value pairs |

**Example Request**

```bash
curl -X POST \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{"site_name":"My Site","site_description":"My Description"}' \
  http://localhost:9002/control-panel/options
```

**Example Response (Success)**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": {
    "message": "Options saved"
  }
}
```

## Authenticated User Endpoints

> All endpoints below require the `Authorization: Bearer <token>` header with a valid user JWT.

### Get User List

```
GET /api/v1/users
```

**Example Request**

```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  http://localhost:9002/api/v1/users
```

**Example Response**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": [
    { "id": 1, "name": "Alice", "email": "alice@example.com" },
    { "id": 2, "name": "Bob", "email": "bob@example.com" }
  ]
}
```

### Create User

```
POST /api/v1/users
```

**Request Body**

| Parameter | Type | Required | Description |
|------|------|------|------|
| name | string | Yes | Username, minimum 2 characters |
| email | string | Yes | Email, must be valid format |

**Example Request**

```bash
curl -X POST \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com"}' \
  http://localhost:9002/api/v1/users
```

**Example Response (Success)**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": {
    "name": "Charlie",
    "email": "charlie@example.com"
  }
}
```

### Get User Details

```
GET /api/v1/users/:id
```

**Path Parameters**

| Parameter | Type | Description |
|------|------|------|
| id | string | User ID |

**Example Request**

```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  http://localhost:9002/api/v1/users/1
```

**Example Response**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": { "id": 1, "name": "Alice", "email": "alice@example.com" }
}
```

**Example Response (Not Found)**

```json
{
  "code": 404,
  "message": "User not found"
}
```

## Authentication

### Token Types

| Token | Duration | Purpose |
|------|------|------|
| Access Token (JWT) | 15 minutes | Authenticate API requests |
| Refresh Token | 7 days | Generate new access tokens |

### Token Flow

1. **Register** or **Login** → Server returns `token` (access) + `refresh_token`
2. Use `token` in `Authorization: Bearer <token>` header for all authenticated requests
3. When `token` expires → Call `/auth/refresh` with `refresh_token` to get a new pair
4. **Admin only**: Changing password increments `token_version`, invalidating all existing tokens

### Unauthenticated Response

```json
{
  "code": 401,
  "message": "Unauthorized"
}
```

## CORS

The backend has CORS middleware enabled, allowing the following cross-origin requests:

- Allowed Origins: All (`*`)
- Allowed Methods: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`
- Allowed Headers: `Origin`, `Content-Type`, `Accept`, `Authorization`

## Multi-Language Support

The API supports switching response languages via the `Accept-Language` header.

```bash
# Chinese (default)
curl http://localhost:9002/api/v1/ping

# English
curl -H "Accept-Language: en" http://localhost:9002/api/v1/ping
```

In Chinese mode, responses return `"message": "操作成功"`. In English mode, they return `"message": "Operation successful"`.

## Turnstile Integration

Some endpoints may require a Cloudflare Turnstile token when the corresponding options are enabled. See [Turnstile Documentation](./TURNSTILE.md) for configuration details.
