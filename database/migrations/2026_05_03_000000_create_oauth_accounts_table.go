package migrations

import (
	"purecore/core"

	"gorm.io/gorm"
)

func init() {
	core.RegisterMigration("create_oauth_accounts_table", func(db *gorm.DB) error {
		return db.Exec(`
			CREATE TABLE IF NOT EXISTS oauth_accounts (
				id BIGSERIAL PRIMARY KEY,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				deleted_at TIMESTAMP WITH TIME ZONE,
				user_id BIGINT NOT NULL,
				provider VARCHAR(50) NOT NULL,
				provider_id VARCHAR(255) NOT NULL,
				email VARCHAR(255) DEFAULT '',
				name VARCHAR(255) DEFAULT '',
				avatar_url VARCHAR(1000) DEFAULT '',
				access_token TEXT DEFAULT '',
				refresh_token TEXT DEFAULT '',
				token_expiry TIMESTAMP WITH TIME ZONE,
				raw_data TEXT DEFAULT ''
			);

			CREATE INDEX IF NOT EXISTS idx_oauth_accounts_user_id ON oauth_accounts(user_id);
			CREATE INDEX IF NOT EXISTS idx_oauth_accounts_provider ON oauth_accounts(provider);
			CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_provider_id ON oauth_accounts(provider_id);
		`).Error
	})
}
