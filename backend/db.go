package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Content struct {
	ID              string `gorm:"primaryKey" json:"id"`
	OriginalContent string `gorm:"type:text" json:"original_content"` // Content from HTML
	EditedContent   string `gorm:"type:text" json:"edited_content"`   // User-modified content
	IsEdited        bool   `json:"is_edited"`                         // True if user has edited
	UpdatedAt       int64  `json:"updated_at"`
}

func InitDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("content.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate the schema
	db.AutoMigrate(&Content{}, &AICommand{})

	return db, nil
}
