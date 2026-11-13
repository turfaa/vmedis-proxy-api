package time2

import (
	"fmt"
	"time"
)

func Today() (from time.Time, until time.Time) {
	return BeginningOfToday(), EndOfToday()
}

func Day(date string) (from time.Time, until time.Time, err error) {
	from, err = BeginningOfDate(date)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("beginning of date: %w", err)
	}

	until, err = EndOfDate(date)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("end of date: %w", err)
	}

	return from, until, nil
}

func BeginningOfToday() time.Time {
	year, month, day := time.Now().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}

func EndOfToday() time.Time {
	year, month, day := time.Now().Date()
	return time.Date(year, month, day, 23, 59, 59, 999999999, time.Local)
}

func BeginningOfDate(date string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date: %w", err)
	}

	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local), nil
}

func EndOfDate(date string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date: %w", err)
	}

	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.Local), nil
}

// BeginningOfLastMonth returns the first instant of the month before the
// current one, in the local timezone.
func BeginningOfLastMonth() time.Time {
	return beginningOfMonthBefore(time.Now())
}

// EndOfLastMonth returns the last instant of the month before the current
// one, in the local timezone.
func EndOfLastMonth() time.Time {
	return endOfMonthBefore(time.Now())
}

// beginningOfMonthBefore steps back from the first day of now's month rather
// than from now itself: going a month back from e.g. December 31 gives
// "November 31", which time.Date normalizes to December 1.
func beginningOfMonthBefore(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.Local)
}

func endOfMonthBefore(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local).Add(-time.Nanosecond)
}
