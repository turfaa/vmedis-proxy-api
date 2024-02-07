package vmedis

import (
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	dateFormat = "02 Jan 2006"
)

// Date is a wrapper around time.Time that implements the Unmarshaler interface
// It only stores the date part of the time.
type Date struct {
	time.Time
}

// UnmarshalDataColumn implements DataColumnUnmarshaler.
func (d *Date) UnmarshalDataColumn(selection *goquery.Selection) error {
	dateString := strings.TrimSpace(selection.Text())

	t, err := time.ParseInLocation(dateFormat, dateString, time.Local)
	if err != nil {
		return err
	}

	d.Time = t
	return nil
}
