package models

import (
	"purecore/core"
	"time"
)

// UserSession represents an active user login session
type UserSession struct {
	core.Model
	UserID       uint      `gorm:"index;not null" json:"user_id"`
	IPAddress    string    `gorm:"type:varchar(45);not null" json:"ip_address"`
	UserAgent    string    `gorm:"type:text" json:"user_agent"`
	DeviceType   string    `gorm:"type:varchar(20);default:''" json:"device_type"`
	DeviceBrand  string    `gorm:"type:varchar(50);default:''" json:"device_brand"`
	DeviceModel  string    `gorm:"type:varchar(50);default:''" json:"device_model"`
	Browser      string    `gorm:"type:varchar(50);default:''" json:"browser"`
	OS           string    `gorm:"type:varchar(50);default:''" json:"os"`
	SessionToken string    `gorm:"type:varchar(255);index;not null" json:"session_token"`
	IsCurrent     bool      `gorm:"default:true" json:"is_current"`
	LoginProvider string    `gorm:"type:varchar(50);default:''" json:"login_provider"`
	LastActivity  time.Time `json:"last_activity"`
	ExpiresAt    time.Time `json:"expires_at"`
}
