package models

import "time"

type Shift struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time

	VmedisID            int       `gorm:"unique"`
	Code                string    `gorm:"index"`
	Cashier             string    `gorm:"index"`
	StartedAt           time.Time `gorm:"index:idx_shift_started_at_ended_at"`
	EndedAt             time.Time `gorm:"index:idx_shift_started_at_ended_at"`
	InitialCash         float64
	ExpectedFinalCash   float64
	ActualFinalCash     float64
	FinalCashDifference float64
	Supervisor          string
	Notes               string
}
