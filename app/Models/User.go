package models

import (
	"purecore/core"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a regular user in the database
type User struct {
	core.Model
	Name            string     `gorm:"type:varchar(100);not null" json:"name" validate:"required,min=2"`
	Email           string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"email" validate:"required,email"`
	Password        string     `gorm:"type:varchar(255);not null;default:''" json:"-"`
	RefreshToken       string     `gorm:"type:varchar(255);default:''" json:"-"`
	RefreshTokenExpiry *time.Time `gorm:"index" json:"-"`
	TokenVersion       int        `gorm:"default:0" json:"-"`
	Avatar          string     `gorm:"type:varchar(500);default:''" json:"avatar"`
	Bio             string     `gorm:"type:text;default:''" json:"bio"`
	Status          string     `gorm:"type:varchar(20);default:'active';index" json:"status"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	LastLoginAt     *time.Time `json:"last_login_at"`
}

// SetPassword hashes the password using bcrypt and stores it
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return nil
}

// CheckPassword verifies the given password against the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
