package dumper

import (
	"context"

	"github.com/turfaa/vmedis-proxy-api/kafkapb"
)

type UpdatedDrugProducer interface {
	ProduceUpdatedDrugByVmedisCode(ctx context.Context, messages []*kafkapb.UpdatedDrugByVmedisCode) error
}
