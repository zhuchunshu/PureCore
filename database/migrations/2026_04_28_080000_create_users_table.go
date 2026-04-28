package migrations

import (
	"purecore/core"

	"gorm.io/gorm"
)

func init() {
	core.RegisterMigration("2026_04_28_080000_create_users_table", upCreateUsersTable)
}

func upCreateUsersTable(db *gorm.DB) error {
	type User struct {
		core.Model
		Name  string `gorm:"type:varchar(100);not null"`
		Email string `gorm:"type:varchar(100);uniqueIndex;not null"`
	}
	return db.AutoMigrate(&User{})
}
