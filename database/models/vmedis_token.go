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

	// Version is the vmedis version the token's session uses, detected when
	// the token is refreshed. It is empty until then.
	Version TokenVersion `gorm:"not null;default:''"`
}

// TokenVersion is the vmedis version a session uses. A session can only open
// the pages of its own version: vmedis redirects it from the other version's
// pages.
type TokenVersion string

const (
	TokenVersionUnknown TokenVersion = ""
	TokenVersionV1      TokenVersion = "V1"
	TokenVersionV2      TokenVersion = "V2"
)

func (v TokenVersion) String() string {
	return string(v)
}

// TokenStatus is the result of checking a token against vmedis.
type TokenStatus struct {
	State   TokenState
	Version TokenVersion
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
