package procurement

import (
	"context"

	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/kafkapb"
)

type UpdatedDrugProducer interface {
	ProduceUpdatedDrugByVmedisCode(ctx context.Context, messages []*kafkapb.UpdatedDrugByVmedisCode) error
}

type DrugUnitsGetter interface {
	GetDrugUnitsByDrugVmedisCodes(ctx context.Context, drugVmedisCodes []string) (map[string][]drug.Unit, error)
}
