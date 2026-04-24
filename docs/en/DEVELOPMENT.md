# PureCore Development Guide

## Development Environment Setup

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Bun 1.0+ (or Node.js 20 + npm)
- Git

### Project Initialization

```bash
# Clone the project
git clone https://github.com/zhuchunshu/PureCore.git
cd PureCore

# Configure environment variables
cp .env.example .env

# Install backend dependencies
go mod tidy

# Create symlink for shared language files
ln -s ../../lang web/public/lang

# Install frontend dependencies
cd web && bun install
```

### Start Development Servers

```bash
# Terminal 1 - Start backend (default port 9002)
go run main.go

# Terminal 2 - Start frontend (default port 9001)
cd web && bun run dev
```

## Project Architecture

### Core Framework (core/)

#### router.go - Route Management

Provides Laravel-style route definitions with chainable calls:

```go
// Create a router group
r := core.NewRouter(app)

// Public routes
r.Prefix("/api/v1").Group(func(r *core.Router) {
    r.Get("/ping", handler)
})

// Authenticated route group
r.Prefix("/api/v1").Middleware(authMiddleware).Group(func(r *core.Router) {
    r.Get("/users", userHandler)
    r.Post("/users", userHandler)
})
```

Supported methods: `Get`, `Post`, `Put`, `Delete`, `Patch`

#### request.go - Request Handling

```go
// Get a single input field (auto-merges param/query/body)
name := req.Input("name")
name := req.Input("name", "default value")

// Get all inputs
allInput := req.All()

// Struct binding and validation
var body CreateUserRequest
if err := req.Validate(&body); err != nil {
    // Handle validation failure
}

// Get authenticated user
user := req.User()

// Get Bearer Token
token := req.BearerToken()

// Get request header
header := req.Header("Content-Type")

// Get client IP
ip := req.IP()
```

#### response.go - Response Handling

```go
// Success response
res.Success(data)                    // 200, code=0

// Error response
res.Error("Error message")           // 400
res.Error("Error message", 422)      // Custom status code

// Unauthorized
res.Unauthorized()                   // 401

// Resource not found
res.NotFound()                       // 404
res.NotFound("User not found")       // Custom message

// Paginated response
res.Paginate(data, total, page, perPage)

// Custom JSON
res.JSON(status, code, message, data)
```

Response messages are automatically translated based on the `Accept-Language` request header.

#### middleware.go - Middleware Bridge

```go
// Controller method signature
type HandlerFunc func(req *Request, res *Response) error

// H() converts HandlerFunc to fiber.Handler
r.Get("/path", core.H(func(req *core.Request, res *core.Response) error {
    return res.Success("hello")
}))

// Or bind to a Controller method
r.Get("/users", core.H(userCtrl.Index))
```

#### lang.go - Multi-Language Manager

```go
// Initialize language files (called in main.go)
core.InitLang("lang")

// Get translation
msg := core.GetLang().Trans("common.success")

// Set current language
core.GetLang().SetLocale("en")

// Get current language
locale := core.GetLang().GetLocale()
```

### Application Layer (app/Http/)

#### Controllers

Controllers are placed in the `app/Http/Controllers/` directory. The method signature is `func(req *core.Request, res *core.Response) error`:

```go
package controllers

import "purecore/core"

type UserController struct{}

func (uc *UserController) Index(req *core.Request, res *core.Response) error {
    // Return user list
    return res.Success(users)
}

func (uc *UserController) Store(req *core.Request, res *core.Response) error {
    // Validate request data
    var body CreateUserRequest
    if err := req.Validate(&body); err != nil {
        return res.Error(err.Error())
    }
    // Create user
    return res.Success(newUser)
}

func (uc *UserController) Show(req *core.Request, res *core.Response) error {
    id := req.Input("id")
    // Find user
    return res.Success(user)
}
```

#### Middleware

Middleware is placed in the `app/Http/Middleware/` directory:

```go
package middleware

import "github.com/gofiber/fiber/v3"

func MyMiddleware() fiber.Handler {
    return func(c fiber.Ctx) error {
        // Pre-processing
        err := c.Next()
        // Post-processing
        return err
    }
}
```

Using middleware in routes:
```go
r.Prefix("/api/v1").Middleware(middleware.MyMiddleware()).Group(func(r *core.Router) {
    // Protected routes
})
```

### Route Registration (routes/)

Register routes in `routes/api.go`:

```go
package routes

func RegisterAPI(r *core.Router) {
    // Public routes
    r.Prefix("/api/v1").Group(func(r *core.Router) {
        r.Get("/ping", ...)
    })

    // Authenticated routes
    r.Prefix("/api/v1").Middleware(middleware.Auth()).Group(func(r *core.Router) {
        r.Get("/users", ...)
        r.Post("/users", ...)
    })
}
```

Call in `main.go`:
```go
router := core.NewRouter(app)
routes.RegisterAPI(router)
```

## Multi-Language Support

### Translation File Structure

Translation files are located in the `lang/` directory, using JSON format with a nested structure:

```json
{
  "common": {
    "success": "Operation successful",
    "error": "Operation failed",
    "not_found": "Resource not found",
    "unauthorized": "Unauthorized"
  },
  "auth": {
    "login_success": "Login successful",
    "token_invalid": "Invalid token"
  }
}
```

### Accessing Translations

Keys use dot notation: `"common.success"`, `"auth.login_success"`

### Backend Usage

```go
// Automatically detects Accept-Language request header
// Default language is Chinese (zh)

// Manually set language
core.GetLang().SetLocale("en")

// Get translation
msg := core.GetLang().Trans("common.success")
```

### Frontend Usage

```javascript
import { t, setLocale, initI18n } from './i18n'

// Initialize (auto-detects browser language)
await initI18n()

// Get translation
t('common.success')  // Chinese: "操作成功", English: "Operation successful"

// Switch language
await setLocale('en')
```

### Adding a New Language

1. Create a new JSON file in the `lang/` directory (e.g., `ja.json`)
2. Fill in translations following the existing file format
3. Both frontend and backend will automatically load it - no additional configuration needed

## Configuration Management

### .env File

```env
# Frontend
FRONTEND_PORT=9001

# Backend
BACKEND_PORT=9002

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=purecore
DB_SSLMODE=disable

# Application
APP_ENV=local
APP_DEBUG=true
```

### Vite Configuration

The frontend `web/vite.config.js` reads `.env` configuration:

```javascript
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  return {
    server: {
      port: parseInt(env.FRONTEND_PORT) || 9001,
      proxy: {
        '/api': {
          target: `http://localhost:${env.BACKEND_PORT || 9002}`,
          changeOrigin: true,
        },
      },
    },
  }
})
```

## Database Configuration

The project uses PostgreSQL. Configure the database connection in the `.env` file.

```go
// Database connection example (using GORM)
import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
    os.Getenv("DB_HOST"),
    os.Getenv("DB_USER"),
    os.Getenv("DB_PASSWORD"),
    os.Getenv("DB_NAME"),
    os.Getenv("DB_PORT"),
    os.Getenv("DB_SSLMODE"),
)

db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
```

## Deployment

### Build Go Backend

```bash
go build -o purecore-server .
./purecore-server
```

### Build Frontend

```bash
cd web
bun run build
# Static files output to web/dist/
```

## FAQ

### 1. Port Conflict

Modify the `FRONTEND_PORT` and `BACKEND_PORT` in the `.env` file.

### 2. Cannot Find lang/ Files

Ensure the symlink has been created:
```bash
ln -s ../../lang web/public/lang
```

### 3. Go Module Issues

```bash
go clean -modcache
go mod tidy
