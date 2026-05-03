package models

import (
	"purecore/core"
	"time"
)

// OAuthAccount links a PureCore user to an OAuth provider identity.
// A single user can have multiple OAuth accounts (e.g., both GitHub and Google).
// A single OAuth provider ID can only be linked to one PureCore user.
type OAuthAccount struct {
	core.Model
	UserID       uint       `gorm:"index;not null" json:"user_id"`
	Provider     string     `gorm:"type:varchar(50);index;not null" json:"provider"`
	ProviderID   string     `gorm:"type:varchar(255);not null;uniqueIndex:idx_oauth_provider_id" json:"provider_id"`
	Email        string     `gorm:"type:varchar(255)" json:"email"`
	Name         string     `gorm:"type:varchar(255)" json:"name"`
	AvatarURL    string     `gorm:"type:varchar(1000)" json:"avatar_url"`
	AccessToken  string     `gorm:"type:text" json:"-"`
	RefreshToken string     `gorm:"type:text" json:"-"`
	TokenExpiry  *time.Time `json:"token_expiry"`
	RawData      string     `gorm:"type:text" json:"-"`
}

// TableName overrides the default table name
func (OAuthAccount) TableName() string {
	return "oauth_accounts"
}
