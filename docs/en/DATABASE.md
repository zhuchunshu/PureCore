# Database & Models

PureCore uses **[GORM](https://gorm.io/)** as its ORM, providing an Eloquent-like experience for database operations.

## Configuration

Database settings are configured via environment variables in your `.env` file:

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host address | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database username | `postgres` |
| `DB_PASSWORD` | Database password | `postgres` |
| `DB_NAME` | Database name | `purecore` |
| `DB_SSLMODE` | SSL mode | `disable` |

## Base Model

All database models should embed `core.Model` to get standard fields:

```go
// core/model.go
type Model struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

## Creating Models

Models live in the `app/Models/` directory. Here's an example:

```go
// app/Models/User.go
package models

import "purecore/core"

type User struct {
    core.Model
    Name  string `gorm:"type:varchar(100);not null" json:"name" validate:"required,min=2"`
    Email string `gorm:"type:varchar(100);uniqueIndex;not null" json:"email" validate:"required,email"`
}
```

**GORM Struct Tags Reference:**

| Tag | Purpose |
|-----|---------|
| `primarykey` | Designates the primary key |
| `type:varchar(n)` | Sets the column type and length |
| `not null` | Disallows NULL values |
| `uniqueIndex` | Creates a unique index |
| `default:value` | Sets a default value |
| `autoIncrement` | Auto-incrementing integer |

## Accessing the Database

Use `core.DB()` anywhere in your application to access the singleton database connection:

```go
import (
    models "purecore/app/Models"
    "purecore/core"
)

// Create
user := models.User{Name: "Alice", Email: "alice@example.com"}
core.DB().Create(&user)

// Read (single)
var user models.User
core.DB().First(&user, 1)        // by primary key
core.DB().Where("email = ?", "alice@example.com").First(&user)

// Read (list)
var users []models.User
core.DB().Find(&users)
core.DB().Where("name LIKE ?", "%Ali%").Find(&users)

// Update
core.DB().Model(&user).Update("name", "Alice Updated")
core.DB().Model(&user).Updates(models.User{Name: "New Name", Email: "new@example.com"})

// Delete (soft delete)
core.DB().Delete(&user)
```

## Migrations

PureCore uses a migration system similar to Laravel's, with each migration registered via `init()` and tracked in the database. Migrations run automatically on server startup and can also be triggered manually:

```bash
./purecore migrate
```

**How it works:**
- Each migration file in `database/migrations/` registers itself via `init()` using `core.RegisterMigration(name, upFunc)`
- `core.RunMigrations()` (in `core/migrator.go`) reads all registered migrations and runs any that haven't been executed
- A `migrations` table in the database tracks which migrations have run and their batch number
- Migrations are compiled into the binary — no filesystem scanning needed
- The `cmd/serve.go` file imports the migrations package (`_ "purecore/database/migrations"`), so all registered migrations are available

**Example migration file:**

```go
// database/migrations/2026_04_27_175200_create_web_options_table.go
package migrations

import (
    "purecore/core"
    "gorm.io/gorm"
)

func init() {
    core.RegisterMigration("2026_04_27_175200_create_web_options_table", up2026_04_27_175200)
}

func up2026_04_27_175200(db *gorm.DB) error {
    type WebOption struct {
        core.Model
        Name  string `gorm:"type:varchar(100);uniqueIndex;not null"`
        Value string `gorm:"type:text;not null;default:''"`
    }
    return db.AutoMigrate(&WebOption{})
}
```

**Creating a migration with the CLI:**

```bash
./purecore make:migration create_posts_table
```

This generates a migration file with the proper `init()` registration and `up()` function. The migration is automatically registered — just rebuild and run `./purecore migrate`.

**Migration naming conventions:**
- Use timestamp prefixes for ordering: `YYYY_MM_DD_HHMMSS_description.go`
- Example: `2026_04_30_120000_create_posts_table.go`

**Migration tracking:**
- The `migrations` table stores migration name, batch number, and timestamp
- Each `migrate` run creates a new batch number for all executed migrations
- Previously-run migrations are skipped via lookup in the tracking table

## Query Building

GORM provides a fluent query builder:

```go
// Where conditions
db.Where("age > ?", 18).Find(&users)
db.Where("name = ? AND email = ?", name, email).Find(&users)

// Ordering
db.Order("created_at desc").Find(&users)

// Limit and Offset
db.Limit(10).Offset(20).Find(&users)

// Count
var count int64
db.Model(&models.User{}).Where("active = ?", true).Count(&count)

// Transactions
db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&user1).Error; err != nil {
        return err
    }
    if err := tx.Create(&user2).Error; err != nil {
        return err
    }
    return nil
})
```

## Debugging

Enable SQL query logging by setting `APP_DEBUG=true` in your `.env` file. This will print all SQL queries to the console.
