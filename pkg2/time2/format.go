package time2

import (
	"time"

	"github.com/klauspost/lctime"
)

func FormatDateTime(t time.Time) string {
	return lctime.Strftime("%d %B %Y, %H:%M:%S %Z", t)
}

func FormatDate(t time.Time) string {
	return lctime.Strftime("%d %B %Y", t)
}
