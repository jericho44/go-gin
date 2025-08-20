package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey" db:"id"`
	Name      string    `json:"name" validate:"required,min=2,max=100" gorm:"not null" db:"name"`
	Email     string    `json:"email" validate:"required,email" gorm:"unique;not null" db:"email"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}
