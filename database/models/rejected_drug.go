package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// RejectedDrug is a drug that was asked by a customer but is not sold (yet).
type RejectedDrug struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	DrugName        string `gorm:"index;not null"`
	Quantity        float64
	Unit            string
	Reason          string
	Resolution      RejectedDrugResolution `gorm:"index;not null;default:UNRESOLVED"`
	ResolutionNotes string
	ResolvedAt      *time.Time `gorm:"index"`
	CreatedBy       string     `gorm:"index"`
	ResolvedBy      string     `gorm:"index"`
}

type RejectedDrugResolution string

const (
	RejectedDrugResolutionUnresolved   RejectedDrugResolution = "UNRESOLVED"
	RejectedDrugResolutionOrdered      RejectedDrugResolution = "ORDERED"
	RejectedDrugResolutionStocked      RejectedDrugResolution = "STOCKED"
	RejectedDrugResolutionSubstituted  RejectedDrugResolution = "SUBSTITUTED"
	RejectedDrugResolutionWillNotStock RejectedDrugResolution = "WILL_NOT_STOCK"
)

// AllRejectedDrugResolutions returns all known resolutions.
func AllRejectedDrugResolutions() []RejectedDrugResolution {
	return []RejectedDrugResolution{
		RejectedDrugResolutionUnresolved,
		RejectedDrugResolutionOrdered,
		RejectedDrugResolutionStocked,
		RejectedDrugResolutionSubstituted,
		RejectedDrugResolutionWillNotStock,
	}
}

func (r RejectedDrugResolution) Valid() bool {
	for _, resolution := range AllRejectedDrugResolutions() {
		if r == resolution {
			return true
		}
	}

	return false
}

func (r *RejectedDrugResolution) Scan(src any) error {
	switch val := src.(type) {
	case string:
		*r = RejectedDrugResolution(val)
	case []byte:
		*r = RejectedDrugResolution(val)
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}

	return nil
}

func (r RejectedDrugResolution) Value() (driver.Value, error) {
	return string(r), nil
}

func (r RejectedDrugResolution) String() string {
	return string(r)
}
