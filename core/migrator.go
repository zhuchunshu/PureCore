package core

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// Migration represents a record in the migrations tracking table
type Migration struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Migration string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"migration"`
	Batch     int       `gorm:"not null;default:1" json:"batch"`
	CreatedAt time.Time `json:"created_at"`
}

// MigrationFunc is the signature for migration functions
type MigrationFunc func(*gorm.DB) error

// registeredMigrations stores migration functions keyed by name
// Registered via init() in migration files
var registeredMigrations = make(map[string]MigrationFunc)

// RegisterMigration registers a migration's up function for later execution.
// Call this from init() in migration files.
func RegisterMigration(name string, up MigrationFunc) {
	registeredMigrations[name] = up
}

// RunMigrations runs all registered migrations that have not yet been executed.
// Migration tracking is done via the migrations table in the database.
// This eliminates the need for filesystem scanning — all migrations are
// compiled into the binary via init() registration and the import in cmd/serve.go.
func RunMigrations() {
	db := DB()

	// Ensure the migrations table exists
	if err := db.AutoMigrate(&Migration{}); err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// Get already-run migrations
	var ranMigrations []Migration
	db.Order("id asc").Find(&ranMigrations)
	ranSet := make(map[string]bool)
	for _, m := range ranMigrations {
		ranSet[m.Migration] = true
	}

	// Determine the next batch number
	nextBatch := 1
	if len(ranMigrations) > 0 {
		nextBatch = ranMigrations[len(ranMigrations)-1].Batch + 1
	}

	// Collect pending migrations from the registered map
	type pendingMigration struct {
		name string
		fn   MigrationFunc
	}
	var pending []pendingMigration
	for name, fn := range registeredMigrations {
		if !ranSet[name] {
			pending = append(pending, pendingMigration{name: name, fn: fn})
		}
	}

	if len(pending) == 0 {
		log.Println("ℹ No pending migrations")
		return
	}

	log.Printf("Running %d pending migration(s)...", len(pending))
	for _, pm := range pending {
		if err := pm.fn(db); err != nil {
			log.Fatalf("Failed to run migration %s: %v", pm.name, err)
		}

		// Record the migration
		if err := db.Create(&Migration{
			Migration: pm.name,
			Batch:     nextBatch,
		}).Error; err != nil {
			log.Fatalf("Failed to record migration %s: %v", pm.name, err)
		}

		log.Printf("  ✓ %s", pm.name)
	}

	log.Printf("✓ %d migration(s) completed (batch %d)", len(pending), nextBatch)
}
