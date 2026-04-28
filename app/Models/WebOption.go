package models

import "purecore/core"

// WebOption represents a key-value site setting stored in the database
type WebOption struct {
	core.Model
	Name  string `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Value string `gorm:"type:text;not null;default:''" json:"value"`
}
