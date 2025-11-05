package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Todo represents a todo item with title and content
type Todo struct {
	ID        string    `gorm:"primaryKey;type:uuid" json:"id"`
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Content   string    `gorm:"type:text" json:"content"`
	Completed bool      `gorm:"default:false" json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate hook to generate UUID
func (t *Todo) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
