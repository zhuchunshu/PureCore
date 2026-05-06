package migrations

import (
	"purecore/core"
	"time"

	"gorm.io/gorm"
)

func init() {
	core.RegisterMigration("2026_05_06_000001_add_refresh_token_expiry", addRefreshTokenExpiry)
}

func addRefreshTokenExpiry(db *gorm.DB) error {
	type User struct {
		RefreshTokenExpiry *time.Time `gorm:"index"`
	}
	type AdminUser struct {
		RefreshTokenExpiry *time.Time `gorm:"index"`
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		return err
	}
	return db.AutoMigrate(&AdminUser{})
}
