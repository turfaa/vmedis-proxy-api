package vmedis

import (
	"fmt"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	timeFormat = "02 Jan 2006 15:04:05"
)

// Time is a wrapper around time.Time that implements the Unmarshaler interface.
type Time struct {
	time.Time
}

// UnmarshalDataColumn implements DataColumnUnmarshaler.
func (t *Time) UnmarshalDataColumn(selection *goquery.Selection) error {
	timeString := selection.Text()

	tt, err := time.ParseInLocation(timeFormat, timeString, time.Local)
	if err != nil {
		return fmt.Errorf("parse time from string [%s]: %w", timeString, err)
	}

	t.Time = tt
	return nil
}
