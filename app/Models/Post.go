package models

import "purecore/core"

// Post represents a post record in the database
type Post struct {
	core.Model
	// Add your fields here
	Name string `gorm:"type:varchar(100);not null" json:"name"`
}
