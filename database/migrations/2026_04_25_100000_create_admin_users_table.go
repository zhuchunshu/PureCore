package migrations

import (
	"purecore/core"

	"gorm.io/gorm"
)

func init() {
	core.RegisterMigration("2026_04_25_100000_create_admin_users_table", upAdminUsers)
}

func upAdminUsers(db *gorm.DB) error {
	type AdminUser struct {
		core.Model
		Username string `gorm:"type:varchar(100);uniqueIndex;not null"`
		Password string `gorm:"type:varchar(255);not null"`
		Name     string `gorm:"type:varchar(100);not null"`
		Role     string `gorm:"type:varchar(50);default:'admin'"`
	}
	return db.AutoMigrate(&AdminUser{})
}
