package sale

import (
	"context"

	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/kafkapb"
)

type DrugsGetter interface {
	GetDrugsByVmedisCodes(ctx context.Context, vmedisCodes []string) ([]drug.Drug, error)
}

type UpdatedDrugProducer interface {
	ProduceUpdatedDrugByVmedisCode(ctx context.Context, messages []*kafkapb.UpdatedDrugByVmedisCode) error
}
