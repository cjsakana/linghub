package model

import "time"

// Message represents a single message in a conversation (user or assistant).
type Message struct {
	ID             uint   `gorm:"primaryKey"`
	ConversationID uint   `gorm:"index;not null"`
	Role           string `gorm:"not null"` // "user" or "assistant"
	Content        string `gorm:"type:text;not null"`
	CreatedAt      time.Time
}

// TableName specifies the database table name for Message.
func (Message) TableName() string {
	return "messages"
}
