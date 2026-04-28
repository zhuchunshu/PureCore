package migrations

import (
	"purecore/core"

	"gorm.io/gorm"
)

func init() {
	core.RegisterMigration("2026_04_27_175200_create_web_options_table", up2026_04_27_175200)
}

func up2026_04_27_175200(db *gorm.DB) error {
	type WebOption struct {
		core.Model
		Name  string `gorm:"type:varchar(100);uniqueIndex;not null"`
		Value string `gorm:"type:text;not null;default:''"`
	}
	return db.AutoMigrate(&WebOption{})
}
