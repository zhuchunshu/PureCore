package models

import (
	"purecore/core"

	"golang.org/x/crypto/bcrypt"
)

// AdminUser represents an administrator in the database
type AdminUser struct {
	core.Model
	Username string `gorm:"type:varchar(100);uniqueIndex;not null" json:"username" validate:"required,min=3"`
	Password string `gorm:"type:varchar(255);not null" json:"-"`
	Name     string `gorm:"type:varchar(100);not null" json:"name"`
	Role     string `gorm:"type:varchar(50);default:'admin'" json:"role"`
}

// SetPassword hashes the password using bcrypt and stores it
func (u *AdminUser) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return nil
}

// CheckPassword verifies the given password against the stored hash
func (u *AdminUser) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
