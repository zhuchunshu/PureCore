package models

import "purecore/core"

// UserOauthProvider stores OAuth platform bindings for users
type UserOauthProvider struct {
	core.Model
	UserID           uint   `gorm:"index;not null" json:"user_id"`
	Provider         string `gorm:"type:varchar(50);index;not null" json:"provider"`
	ProviderUserID   string `gorm:"type:varchar(255);not null" json:"provider_user_id"`
	ProviderEmail    string `gorm:"type:varchar(100);default:''" json:"provider_email"`
	ProviderUsername string `gorm:"type:varchar(100);default:''" json:"provider_username"`
	AccessToken      string `gorm:"type:text;default:''" json:"-"`
	RefreshToken     string `gorm:"type:text;default:''" json:"-"`
	TokenExpiresAt   string `gorm:"type:varchar(50);default:''" json:"token_expires_at"`
	AvatarURL        string `gorm:"type:varchar(500);default:''" json:"avatar_url"`
	RawData          string `gorm:"type:text;default:''" json:"-"`
}
