package dumper

import (
	"context"
	"log"

	"github.com/turfaa/vmedis-proxy-api/drug"
)

// DumpDrugs dumps the drugs.
func DumpDrugs(ctx context.Context, drugService *drug.Service) {
	if err := drugService.DumpDrugsFromVmedisToDB(ctx); err != nil {
		log.Println("Error dumping drugs:", err)
	}
}
