package time2

import (
	"testing"
	"time"
)

func TestMonthBefore(t *testing.T) {
	tests := []struct {
		now       string
		wantBegin string
		wantEnd   string
	}{
		{now: "2026-11-14", wantBegin: "2026-10-01", wantEnd: "2026-10-31"},
		{now: "2026-11-01", wantBegin: "2026-10-01", wantEnd: "2026-10-31"},
		// The previous month is shorter than today's day of the month.
		{now: "2026-12-31", wantBegin: "2026-11-01", wantEnd: "2026-11-30"},
		{now: "2026-10-31", wantBegin: "2026-09-01", wantEnd: "2026-09-30"},
		{now: "2026-03-29", wantBegin: "2026-02-01", wantEnd: "2026-02-28"},
		{now: "2028-03-31", wantBegin: "2028-02-01", wantEnd: "2028-02-29"},
		// Across the year boundary.
		{now: "2027-01-31", wantBegin: "2026-12-01", wantEnd: "2026-12-31"},
	}

	for _, tt := range tests {
		t.Run(tt.now, func(t *testing.T) {
			date, err := time.ParseInLocation(time.DateOnly, tt.now, time.Local)
			if err != nil {
				t.Fatal(err)
			}
			now := date.Add(23*time.Hour + 30*time.Minute)

			begin, end := beginningOfMonthBefore(now), endOfMonthBefore(now)

			wantBegin, _ := time.ParseInLocation(time.DateOnly, tt.wantBegin, time.Local)
			if !begin.Equal(wantBegin) {
				t.Errorf("beginningOfMonthBefore(%s) = %s, want %s", now, begin, wantBegin)
			}

			wantEndDay, _ := time.ParseInLocation(time.DateOnly, tt.wantEnd, time.Local)
			wantEnd := wantEndDay.AddDate(0, 0, 1).Add(-time.Nanosecond)
			if !end.Equal(wantEnd) {
				t.Errorf("endOfMonthBefore(%s) = %s, want %s", now, end, wantEnd)
			}
		})
	}
}
