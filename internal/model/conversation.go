package model

import "time"

// Conversation represents a chat conversation belonging to a user.
type Conversation struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"index;not null"`
	Title     string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName specifies the database table name for Conversation.
func (Conversation) TableName() string {
	return "conversations"
}
