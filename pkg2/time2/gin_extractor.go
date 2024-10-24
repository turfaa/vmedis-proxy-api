package time2

import (
	"time"

	"github.com/gin-gonic/gin"
)

func GetOneDayFromQuery(c *gin.Context) (from time.Time, until time.Time, err error) {
	return ParseTimeRange(c.Query("date"), "", "")
}

func GetTimeRangeFromQuery(c *gin.Context) (from time.Time, until time.Time, err error) {
	untilQuery := c.Query("until")
	if untilQuery == "" {
		untilQuery = c.Query("to")
	}
	return ParseTimeRange(c.Query("date"), c.Query("from"), untilQuery)
}
