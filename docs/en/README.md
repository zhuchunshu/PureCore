# PureCore Framework

## Introduction

PureCore is a full-stack Go web development framework that wraps GoFiber v3 into a Laravel-like development style. It provides routing groups, middleware pipelines, request validation, unified response formatting, multi-language support, and many more out-of-the-box features. The frontend uses Vue 3 + Vite + Tailwind CSS + DaisyUI.

- GitHub: [https://github.com/zhuchunshu/PureCore](https://github.com/zhuchunshu/PureCore)

## Tech Stack

| Layer | Technology |
|------|------|
| Backend Language | Go |
| Backend Framework | GoFiber v3 |
| Database | PostgreSQL |
| ORM | GORM |
| Validation | go-playground/validator |
| Frontend Framework | Vue 3 |
| Build Tool | Vite |
| CSS Framework | Tailwind CSS + DaisyUI |
| Package Manager | Bun |

## Project Structure

```
/purecore
├── core/                  # Core framework
│   ├── router.go          # Laravel-style routing (group/prefix/middleware)
│   ├── request.go         # Request handling (Input/Validate/User)
│   ├── response.go        # Unified responses (Success/Error/Paginate)
│   ├── middleware.go       # HandlerFunc type and H() bridge function
│   └── lang.go            # Multi-language manager
├── app/Http/              # Application layer
│   ├── Controllers/       # Controllers
│   │   └── UserController.go
│   └── Middleware/         # Middleware
│       ├── Auth.go        # Token authentication
│       ├── Cors.go        # CORS handling
│       └── Lang.go        # Language detection
├── routes/                # Route registration
│   └── api.go
├── lang/                  # Translation files (shared frontend & backend)
│   ├── zh.json            # Chinese translations
│   └── en.json            # English translations
├── web/                   # Frontend project
│   ├── src/
│   │   ├── i18n.js        # Frontend i18n module
│   │   ├── App.vue
│   │   └── main.js
│   ├── public/
│   │   └── lang/          # -> ../../lang (symlink)
│   ├── vite.config.js
│   └── package.json
├── .env                   # Environment config (shared)
├── main.go                # Backend entry point
├── go.mod
└── go.sum
```

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL
- Bun (or Node.js + npm)
- Git

### 1. Clone the Project

```bash
git clone https://github.com/zhuchunshu/PureCore.git
cd PureCore
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env file to set database connection and ports
```

### 3. Start Backend

```bash
# Install dependencies
go mod tidy

# Create symlink for shared language files
ln -s ../../lang web/public/lang

# Start server (default port 9002)
go run main.go
```

### 4. Start Frontend

```bash
cd web
bun install
bun run dev
# Frontend runs at http://localhost:9001
# API requests are proxied to http://localhost:9002
```

## Core Features

### Laravel-Style Routing

```go
// Route groups, prefixes, and middleware
r.Prefix("/api/v1").Middleware(middleware.Auth()).Group(func(r *core.Router) {
    r.Get("/users", core.H(userCtrl.Index))
    r.Post("/users", core.H(userCtrl.Store))
    r.Get("/users/:id", core.H(userCtrl.Show))
})
```

### Request Validation

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
}

func (uc *UserController) Store(req *core.Request, res *core.Response) error {
    var body CreateUserRequest
    if err := req.Validate(&body); err != nil {
        return res.Error(err.Error())
    }
    // Handle business logic...
}
```

### Unified Responses

```go
res.Success(data)           // {"code":0,"message":"Operation successful","data":{...}}
res.Error("Error message")  // {"code":400,"message":"Error message"}
res.NotFound("User not found") // {"code":404,"message":"User not found"}
res.Unauthorized()          // {"code":401,"message":"Unauthorized"}
res.Paginate(data, total, page, perPage) // Paginated response
```

### Multi-Language Support

Frontend and backend share the same JSON translation files in the `lang/` directory. Default language is **Chinese**.

- **Backend**: Automatically detects language via `Accept-Language` request header. Response messages use the corresponding language.
- **Frontend**: Automatically detects browser language. Use `t('common.success')` to get translations.

```javascript
// Frontend usage
import { t, setLocale } from './i18n'
console.log(t('common.success'))  // "操作成功" (Chinese)
setLocale('en')
console.log(t('common.success'))  // "Operation successful"
```

To add a new language, simply create a new JSON file in the `lang/` directory.

## Configuration

The project uses a `.env` file for configuration, shared between frontend and backend.

| Variable | Description | Default |
|------|------|--------|
| FRONTEND_PORT | Frontend dev server port | 9001 |
| BACKEND_PORT | Backend API server port | 9002 |
| DB_HOST | Database host address | localhost |
| DB_PORT | Database port | 5432 |
| DB_USER | Database username | postgres |
| DB_PASSWORD | Database password | postgres |
| DB_NAME | Database name | purecore |
| APP_ENV | Runtime environment | local |
| APP_DEBUG | Debug mode | true |

## API Endpoints

For detailed API documentation, see [API Documentation](./API.md).

### Public Endpoints

| Method | Path | Description |
|------|------|------|
| GET | /api/v1/ping | Health check |
| GET | /api/v1/system/info | Project information |

### Authenticated Endpoints

| Method | Path | Description |
|------|------|------|
| GET | /api/v1/users | User list |
| POST | /api/v1/users | Create user |
| GET | /api/v1/users/:id | User details |

Authentication: `Authorization: Bearer <token>`

## Development Guide

For detailed development documentation, see [Development Guide](./DEVELOPMENT.md).
