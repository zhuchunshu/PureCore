package core

import (
	"fmt"
	"log"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	once sync.Once
)

// DB returns the singleton database connection.
// Uses the centralized Config system for all connection parameters.
func DB() *gorm.DB {
	once.Do(func() {
		db = connectDB()
	})
	return db
}

// connectDB establishes a PostgreSQL connection using the Config system.
func connectDB() *gorm.DB {
	cfg := GetConfig()
	host, port, user, password, dbname, sslmode := cfg.DatabaseConfig()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	logLevel := logger.Warn
	if cfg.AppDebug() {
		logLevel = logger.Info
	}

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✓ Database connected")
	return conn
}

// ResetDB closes the existing database connection and resets the singleton.
// Useful for testing or when configuration changes at runtime.
func ResetDB() {
	if db != nil {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
	once = sync.Once{}
	db = nil
}
