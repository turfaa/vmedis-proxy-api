package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type VmedisToken struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	Token string     `gorm:"unique;not null"`
	State TokenState `gorm:"type:token_state;not null;default:UNCHECKED"`
}

type TokenState string

const (
	TokenStatePending TokenState = "UNCHECKED"
	TokenStateActive  TokenState = "ACTIVE"
	TokenStateExpired TokenState = "EXPIRED"
)

func (s *TokenState) Scan(src any) error {
	switch val := src.(type) {
	case string:
		*s = TokenState(val)
	case []byte:
		*s = TokenState(val)
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}

	return nil
}

func (s TokenState) Value() (driver.Value, error) {
	return string(s), nil
}

func (s TokenState) String() string {
	return string(s)
}
