package migrations

import (
	"purecore/core"
	"time"

	"gorm.io/gorm"
)

func init() {
	core.RegisterMigration("2026_05_01_000000_create_user_sessions_table", upCreateUserSessionsTable)
}

func upCreateUserSessionsTable(db *gorm.DB) error {
	type UserSession struct {
		core.Model
		UserID       uint   `gorm:"index;not null"`
		IPAddress    string `gorm:"type:varchar(45);not null"`
		UserAgent    string `gorm:"type:text"`
		DeviceType   string `gorm:"type:varchar(20);default:''"`
		DeviceBrand  string `gorm:"type:varchar(50);default:''"`
		DeviceModel  string `gorm:"type:varchar(50);default:''"`
		Browser      string `gorm:"type:varchar(50);default:''"`
		OS           string `gorm:"type:varchar(50);default:''"`
		SessionToken string `gorm:"type:varchar(255);index;not null"`
		IsCurrent    bool   `gorm:"default:true"`
		LastActivity time.Time
		ExpiresAt    time.Time
	}
	return db.AutoMigrate(&UserSession{})
}
