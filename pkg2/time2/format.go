package time2

import (
	"fmt"
	"time"

	"github.com/klauspost/lctime"
)

// localizer formats dates in Indonesian.
//
// We hold our own localizer instead of calling lctime.Strftime, which formats
// with a package-global locale that lctime initializes from LC_TIME / LC_ALL /
// LANG at import time. When lctime does not recognize that value it installs an
// empty locale rather than falling back to POSIX, and formatting a month name
// then panics on an empty slice. So on a host with, say, LANG=C.UTF-8 (GitHub
// Actions runners) every caller that had not gone through main's
// lctime.SetLocale would crash. Owning the localizer keeps formatting identical
// in the server, in commands, and in tests, whatever the host locale is.
var localizer = mustNewLocalizer("id_ID")

func mustNewLocalizer(id string) lctime.Localizer {
	localizer, err := lctime.NewLocalizer(id)
	if err != nil {
		panic(fmt.Sprintf("time2: load %s locale: %s", id, err))
	}

	return localizer
}

func FormatDateTime(t time.Time) string {
	return localizer.Strftime("%d %B %Y, %H:%M:%S %Z", t)
}

func FormatDate(t time.Time) string {
	return localizer.Strftime("%d %B %Y", t)
}
