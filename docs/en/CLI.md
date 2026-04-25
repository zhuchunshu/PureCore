# CLI Commands

PureCore provides an Artisan-style CLI powered by [Cobra](https://github.com/spf13/cobra).

## Available Commands

| Command | Description |
|---------|-------------|
| `./purecore serve` | Start the HTTP server |
| `./purecore make:model` | Create a new model file |
| `./purecore make:controller` | Create a new controller file |
| `./purecore make:migration` | Create a new migration file |
| `./purecore migrate` | Run database migrations |
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
4. Registers middleware (CORS, language detection)
5. Registers all API routes
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
