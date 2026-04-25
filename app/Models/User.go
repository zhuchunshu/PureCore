package models

import "purecore/core"

// User represents a user in the database
type User struct {
	core.Model
	Name  string `gorm:"type:varchar(100);not null" json:"name" validate:"required,min=2"`
	Email string `gorm:"type:varchar(100);uniqueIndex;not null" json:"email" validate:"required,email"`
}
