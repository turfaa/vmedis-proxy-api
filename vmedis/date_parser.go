package vmedis

import (
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	dateFormat           = "02 Jan 2006"
	dateTimeMinuteFormat = "02 Jan 2006 15:04"
)

// Date is a wrapper around time.Time that implements the Unmarshaler interface
// It only stores the date part of the time.
type Date struct {
	time.Time
}

// UnmarshalDataColumn implements DataColumnUnmarshaler.
func (d *Date) UnmarshalDataColumn(selection *goquery.Selection) error {
	dateBytes := []byte(selection.Text())
	dateBytes = compactSpaces(dateBytes)
	dateString := strings.TrimSpace(string(dateBytes))

	t, err := time.ParseInLocation(dateFormat, dateString, time.Local)
	if err != nil {
		return err
	}

	d.Time = t
	return nil
}
