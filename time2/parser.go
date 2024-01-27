package time2

import (
	"fmt"
	"time"
)

func ParseTimeRange(dateQuery, fromQuery, untilQuery string) (from time.Time, until time.Time, err error) {
	if fromQuery != "" {
		from, _, err = Day(fromQuery)
		if err != nil {
			err = fmt.Errorf("parse time range from `from` query [%s]: %w", fromQuery, err)
			return
		}

		if untilQuery != "" {
			_, until, err = Day(untilQuery)
			if err != nil {
				err = fmt.Errorf("parse time range from `until` query [%s]: %w", untilQuery, err)
				return
			}
		} else {
			until = EndOfToday()
		}
	} else if dateQuery == "" {
		from, until = Today()
	} else {
		from, until, err = Day(dateQuery)
		if err != nil {
			err = fmt.Errorf("parse time range from `date` query [%s]: %w", dateQuery, err)
			return
		}
	}

	return
}
