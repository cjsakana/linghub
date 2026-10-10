package model

import "time"

// User represents a registered user in the system.
type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName specifies the database table name for User.
func (User) TableName() string {
	return "users"
}
