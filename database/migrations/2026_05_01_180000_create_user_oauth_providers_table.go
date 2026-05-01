package migrations

import (
	"purecore/core"

	"gorm.io/gorm"
)

func init() {
	core.RegisterMigration("2026_05_01_180000_create_user_oauth_providers_table", upCreateUserOauthProvidersTable)
}

func upCreateUserOauthProvidersTable(db *gorm.DB) error {
	type UserOauthProvider struct {
		core.Model
		UserID           uint   `gorm:"index;not null"`
		Provider         string `gorm:"type:varchar(50);index;not null"`
		ProviderUserID   string `gorm:"type:varchar(255);not null"`
		ProviderEmail    string `gorm:"type:varchar(100);default:''"`
		ProviderUsername string `gorm:"type:varchar(100);default:''"`
		AccessToken      string `gorm:"type:text;default:''"`
		RefreshToken     string `gorm:"type:text;default:''"`
		TokenExpiresAt   string `gorm:"type:varchar(50);default:''"`
		AvatarURL        string `gorm:"type:varchar(500);default:''"`
		RawData          string `gorm:"type:text;default:''"`
	}

	// Create unique index on provider + provider_user_id
	db.AutoMigrate(&UserOauthProvider{})
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_oauth_providers_provider_user ON user_oauth_providers(provider, provider_user_id)")

	return nil
}
