package migrations

import (
	"purecore/core"

	"gorm.io/gorm"
)

func init() {
	core.RegisterMigration("2026_05_06_000000_add_login_provider_to_user_sessions", upAddLoginProviderToUserSessions)
}

func upAddLoginProviderToUserSessions(db *gorm.DB) error {
	type UserSession struct {
		LoginProvider string `gorm:"type:varchar(50);default:''"`
	}
	return db.Migrator().AddColumn(&UserSession{}, "login_provider")
}
