package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BlogPost represents a blog post for development journal
type BlogPost struct {
	ID        string    `gorm:"primaryKey;type:uuid" json:"id"`
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Tags      string    `gorm:"type:varchar(500)" json:"tags"` // Comma-separated tags
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate hook to generate UUID
func (b *BlogPost) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}
