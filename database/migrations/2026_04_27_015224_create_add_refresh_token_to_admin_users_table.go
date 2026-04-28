package migrations

import (
	"purecore/core"

	"gorm.io/gorm"
)

func init() {
	core.RegisterMigration("2026_04_27_015224_create_add_refresh_token_to_admin_users_table", upAddRefreshTokenToAdminUsers)
}

func upAddRefreshTokenToAdminUsers(db *gorm.DB) error {
	type AdminUser struct {
		core.Model
		Username     string `gorm:"type:varchar(100);uniqueIndex;not null"`
		Password     string `gorm:"type:varchar(255);not null"`
		Name         string `gorm:"type:varchar(100);not null"`
		Role         string `gorm:"type:varchar(50);default:'admin'"`
		RefreshToken string `gorm:"type:varchar(255);default:''"`
		TokenVersion int    `gorm:"default:0"`
	}
	return db.AutoMigrate(&AdminUser{})
}
