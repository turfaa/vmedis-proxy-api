package models

import "time"

// User defins a user.
type User struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Email string `gorm:"unique"`
	Role  string `gorm:"index"`
}
