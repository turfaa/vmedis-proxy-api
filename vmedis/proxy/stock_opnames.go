package proxy

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/vmedis/dumper"
)

// HandleDumpStockOpnames handles the request to dump the stock opnames.
func (s *ApiServer) HandleDumpStockOpnames(c *gin.Context) {
	go dumper.DumpDailyStockOpnames(context.Background(), s.DB, s.Client)
	c.JSON(200, gin.H{
		"message": "dumping stock opnames",
	})
}
