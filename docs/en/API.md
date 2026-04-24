# PureCore API Documentation

## General Information

- Base URL: `/api/v1`
- Default Port: `9002`
- Authentication: `Bearer Token`
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

Returns project metadata from `web/package.json` (under the `purecore` key), including version, release type, author, license, and dependencies.

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

## Authenticated Endpoints (Login Required)

> All endpoints below require the `Authorization: Bearer <token>` header

### Get User List

```
GET /api/v1/users
```

**Request Headers**

```
Authorization: Bearer valid-token
```

**Example Request**

```bash
curl -H "Authorization: Bearer valid-token" http://localhost:9002/api/v1/users
```

**Example Response**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": [
    { "id": "1", "name": "Alice" },
    { "id": "2", "name": "Bob" }
  ]
}
```

### Create User

```
POST /api/v1/users
```

**Request Headers**

```
Authorization: Bearer valid-token
Content-Type: application/json
```

**Request Parameters**

| Parameter | Type | Required | Description |
|------|------|------|------|
| name | string | Yes | Username, minimum 2 characters |
| email | string | Yes | Email, must be valid format |

**Example Request**

```bash
curl -X POST \
  -H "Authorization: Bearer valid-token" \
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

**Example Response (Validation Failed)**

```json
{
  "code": 400,
  "message": "Validation failed"
}
```

### Get User Details

```
GET /api/v1/users/:id
```

**Request Headers**

```
Authorization: Bearer valid-token
```

**Path Parameters**

| Parameter | Type | Description |
|------|------|------|
| id | string | User ID |

**Example Request**

```bash
curl -H "Authorization: Bearer valid-token" http://localhost:9002/api/v1/users/1
```

**Example Response**

```json
{
  "code": 0,
  "message": "Operation successful",
  "data": { "id": "1" }
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

### Getting a Token

For testing, use the fixed token: `valid-token`

In production, implement a full JWT authentication flow:

1. User logs in to receive a token
2. Include `Authorization: Bearer <token>` in request headers
3. Backend validates the token

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
