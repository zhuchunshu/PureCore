# CLI Commands

PureCore provides an Artisan-style CLI powered by [Cobra](https://github.com/spf13/cobra).

## Available Commands

| Command | Description |
|---------|-------------|
| `./purecore serve` | Start the HTTP server |
| `./purecore migrate` | Run database migrations |
| `./purecore make:model` | Create a new model file |
| `./purecore make:controller` | Create a new controller file |
| `./purecore make:migration` | Create a new migration file |
| `./purecore make:module` | Create a new service provider module |
| `./purecore make:middleware` | Create a new middleware file |
| `./purecore --help` | Show all available commands |
| `./purecore completion` | Generate shell autocompletion script |

## serve

Start the PureCore HTTP server.

```bash
./purecore serve
```

The server listens on the port specified by the `BACKEND_PORT` environment variable (default: `9002`).

**What happens on startup:**
1. Loads environment variables from `.env`
2. Initializes the language manager (`lang/` directory)
3. Establishes a database connection (GORM + PostgreSQL)
4. Loads and runs all registered service providers (routes, middleware)
5. Runs any pending database migrations
6. Starts listening for HTTP requests

## migrate

Run all registered database migrations that have not yet been executed.

```bash
./purecore migrate
```

**How it works:**
- Connects to the database using credentials from `.env`
- Creates a `migrations` table if it doesn't exist to track execution history
- Iterates over all migrations registered via `init()` in `database/migrations/`
- Runs GORM's `AutoMigrate` on each pending migration
- Records each migration in the database to prevent re-execution

Migrations are automatically included in the binary via Go's `init()` mechanism — no filesystem scanning needed. The `cmd/serve.go` file imports the migrations package, so all registered migrations are compiled in and run automatically when the server starts.

**Adding a new model using make commands:**

```bash
./purecore make:model Post
./purecore make:migration Post
```

Then rebuild and run:

```bash
go build -o purecore .
./purecore migrate
```

## make:model

Create a new GORM model file in `app/Models/`.

```bash
./purecore make:model Post
```

This generates `app/Models/Post.go` with:
- Package declaration and `purecore/core` import
- A struct embedding `core.Model`
- A `Name` field with GORM and validation tags

**After creating a model:**
1. Add the model to the migration list in `cmd/migrate.go`
2. Run `./purecore migrate` to create the database table

## make:controller

Create a new controller file in `app/Http/Controllers/` with full CRUD scaffold.

```bash
./purecore make:controller Post
```

This generates `app/Http/Controllers/PostController.go` with:
- `Index` — List all records
- `Store` — Create a new record (with validation)
- `Show` — Get a single record by ID

Each method uses the corresponding model from `app/Models/` and accesses the database via `core.DB()`.

## make:migration

Create a new migration file in `database/migrations/`.

```bash
./purecore make:migration Post
```

This generates a migration file with:
- `init()` registration — Calls `core.RegisterMigration()` to register the migration automatically
- `up()` function — Creates the table using GORM AutoMigrate with `core.Model` embedding

The migration is automatically registered when the binary is compiled — no need to manually add it to any list. Simply rebuild and run `./purecore migrate`.

## make:module

Create a new service provider module in `app/Providers/`.

```bash
./purecore make:module Post
```

This generates `app/Providers/PostServiceProvider.go` that implements the `core.ServiceProvider` interface with:
- `Name()` — Returns the provider's unique identifier (`"post"`)
- `Register(router)` — Placeholder for adding routes and middleware
- `Boot()` — Post-registration initialization hook

**After creating a module:**

1. Add the provider to the `AddProviders()` chain in `cmd/serve.go`:
```go
app.AddProviders(
    &providers.RouteServiceProvider{},
    &providers.PostServiceProvider{},  // ← Add your new provider
)
```
2. Implement your routes in the `Register()` method
3. Add any initialization logic in the `Boot()` method

**Example Register() implementation:**

```go
func (p *PostServiceProvider) Register(router *core.Router) error {
    router.Prefix("/api/v1").Group(func(r *core.Router) {
        r.Get("/posts", core.H(postCtrl.Index))
        r.Post("/posts", core.H(postCtrl.Store))
    })
    return nil
}
```

## make:middleware

Create a new middleware file in `app/Http/Middleware/`.

```bash
./purecore make:middleware RateLimit
```

This generates `app/Http/Middleware/RateLimit.go` with a function that returns `fiber.Handler`:

```go
package middleware

import "github.com/gofiber/fiber/v3"

func RateLimit() fiber.Handler {
    return func(c fiber.Ctx) error {
        // Add your middleware logic here
        return c.Next()
    }
}
```

**Using the middleware in routes:**

```go
// Direct use
router.Middleware(middleware.RateLimit()).Group(func(r *core.Router) {
    r.Get("/protected", handler)
})

// Register as named middleware (in a service provider):
router.RegisterNamedMiddlewares(map[string]core.MiddlewareFunc{
    "rate_limit": middleware.RateLimit(),
})

// Then use by name
router.MiddlewareByName("rate_limit").Group(func(r *core.Router) {
    r.Get("/protected", handler)
})
```

**For global middleware**, add it to `RunWithMiddleware()` in `cmd/serve.go`.

## Adding New Commands

1. Create a new file in `cmd/` (e.g., `cmd/mycommand.go`)
2. Register it in `init()` with `rootCmd.AddCommand(mycmd)`
3. Rebuild: `go build -o purecore .`

```go
// cmd/mycommand.go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Description of my command",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Hello from my command!")
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
}
