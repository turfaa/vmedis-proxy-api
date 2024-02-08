package dumper

import (
	"context"
	"log"

	"github.com/turfaa/vmedis-proxy-api/procurement"
)

const (
	procurementRecommendationsKey = "static_key.procurement_recommendations.json.zlib"
)

// DumpProcurementRecommendations calculates and dumps procurement recommendations to cache.
func DumpProcurementRecommendations(ctx context.Context, procurementService *procurement.Service) {
	if err := procurementService.DumpRecommendationsFromVmedisToRedis(ctx); err != nil {
		log.Fatalf("DumpProcurementRecommendations: %s", err)
	}
}
