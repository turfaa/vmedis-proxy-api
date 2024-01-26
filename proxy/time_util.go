package proxy

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func getOneDayFromQuery(c *gin.Context) (from time.Time, until time.Time, err error) {
	return getTimeRange(c.Query("date"), "", "")
}

func getTimeRangeFromQuery(c *gin.Context) (from time.Time, until time.Time, err error) {
	return getTimeRange(c.Query("date"), c.Query("from"), c.Query("until"))
}

func getTimeRange(dateQuery, fromQuery, untilQuery string) (from time.Time, until time.Time, err error) {
	if fromQuery != "" {
		from, _, err = day(fromQuery)
		if err != nil {
			err = fmt.Errorf("parse time range from `from` query [%s]: %w", fromQuery, err)
			return
		}

		if untilQuery != "" {
			_, until, err = day(untilQuery)
			if err != nil {
				err = fmt.Errorf("parse time range from `until` query [%s]: %w", untilQuery, err)
				return
			}
		} else {
			until = endOfToday()
		}
	} else if dateQuery == "" {
		from, until = today()
	} else {
		from, until, err = day(dateQuery)
		if err != nil {
			err = fmt.Errorf("parse time range from `date` query [%s]: %w", dateQuery, err)
			return
		}
	}

	return
}

func today() (from time.Time, until time.Time) {
	return beginningOfToday(), endOfToday()
}

func day(date string) (from time.Time, until time.Time, err error) {
	from, err = beginningOfDate(date)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("beginning of date: %w", err)
	}

	until, err = endOfDate(date)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("end of date: %w", err)
	}

	return from, until, nil
}

func beginningOfToday() time.Time {
	year, month, day := time.Now().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}

func endOfToday() time.Time {
	year, month, day := time.Now().Date()
	return time.Date(year, month, day, 23, 59, 59, 999999999, time.Local)
}

func beginningOfDate(date string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date: %w", err)
	}

	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local), nil
}

func endOfDate(date string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date: %w", err)
	}

	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.Local), nil
}
