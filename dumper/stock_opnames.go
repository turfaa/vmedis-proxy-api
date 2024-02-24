package dumper

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/turfaa/vmedis-proxy-api/stockopname"
)

var (
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// DumpDailyStockOpnames dumps the daily stock opnames.
func DumpDailyStockOpnames(ctx context.Context, stockOpnameService *stockopname.Service) {
	if err := stockOpnameService.DumpTodayStockOpnamesFromVmedisToDB(ctx); err != nil {
		log.Fatalf("DumpTodayStockOpnamesFromVmedisToDB: %s", err)
	}
}
