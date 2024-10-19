package vmedis

import (
	"fmt"
	"slices"
	"strings"
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
	timeBytes := []byte(selection.Text())
	timeBytes = compactSpaces(timeBytes)
	timeString := strings.TrimSpace(string(timeBytes))

	tt, err := time.ParseInLocation(timeFormat, timeString, time.Local)
	if err != nil {
		return fmt.Errorf("parse time from string [%s]: %w", timeString, err)
	}

	t.Time = tt
	return nil
}

func compactSpaces(bytes []byte) []byte {
	bytes = slices.CompactFunc(bytes, func(x, y byte) bool {
		return isSpace(x) && isSpace(y)
	})

	for i := 0; i < len(bytes); i++ {
		if isSpace(bytes[i]) {
			bytes[i] = ' '
		}
	}

	return bytes
}

func isSpace(x byte) bool {
	return x == ' ' || x == '\n' || x == '\t' || x == '\r' || x == '\v' || x == '\f' || x == '\u00a0'
}
