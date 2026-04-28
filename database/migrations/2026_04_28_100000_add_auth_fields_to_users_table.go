package migrations

import (
	"purecore/core"

	"gorm.io/gorm"
)

func init() {
	core.RegisterMigration("2026_04_28_100000_add_auth_fields_to_users_table", upAddAuthFieldsToUsers)
}

func upAddAuthFieldsToUsers(db *gorm.DB) error {
	type User struct {
		Password     string `gorm:"type:varchar(255);not null;default:''"`
		RefreshToken string `gorm:"type:varchar(255);default:''"`
		TokenVersion int    `gorm:"default:0"`
	}
	return db.AutoMigrate(&User{})
}
