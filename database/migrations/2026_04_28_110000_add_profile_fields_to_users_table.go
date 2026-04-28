package migrations

import (
	"purecore/core"
	"time"

	"gorm.io/gorm"
)

func init() {
	core.RegisterMigration("2026_04_28_110000_add_profile_fields_to_users_table", upAddProfileFieldsToUsers)
}

func upAddProfileFieldsToUsers(db *gorm.DB) error {
	type User struct {
		Avatar          string     `gorm:"type:varchar(500);default:''" json:"avatar"`
		Bio             string     `gorm:"type:text;default:''" json:"bio"`
		Status          string     `gorm:"type:varchar(20);default:'active';index" json:"status"`
		EmailVerifiedAt *time.Time `json:"email_verified_at"`
		LastLoginAt     *time.Time `json:"last_login_at"`
	}
	return db.AutoMigrate(&User{})
}
