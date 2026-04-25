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

Run the migrate command to create or update tables based on your models:

```bash
./purecore migrate
```

**How it works:**
- `cmd/migrate.go` contains a list of all models to migrate
- GORM's `AutoMigrate` creates missing tables and columns
- Existing data is preserved — migrations are non-destructive
- New models must be registered in the `models` slice in `cmd/migrate.go`

**Adding a new model to migrations:**

```go
// In cmd/migrate.go
func migrateRun(cmd *cobra.Command, args []string) {
    db := core.DB()

    models := []interface{}{
        &models.User{},
        &models.Post{},    // ← Add your new model here
    }

    for _, model := range models {
        name := reflect.TypeOf(model).Elem().Name()
        if err := db.AutoMigrate(model); err != nil {
            log.Fatalf("Failed to migrate %s: %v", name, err)
        }
        log.Printf("  ✓ %s table migrated", name)
    }
}
```

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
