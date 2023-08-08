package models

import (
	"time"
)

// Drug represents a drug.
type Drug struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	VmedisID     int    `gorm:"index"`
	VmedisCode   string `gorm:"unique"`
	Name         string
	Manufacturer string `gorm:"index"`
}
