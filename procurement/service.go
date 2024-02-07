package procurement

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/kafkapb"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

type Service struct {
	db           *Database
	vmedis       *vmedis.Client
	drugProducer UpdatedDrugProducer
}

func (s *Service) DumpProcurementsBetweenDatesFromVmedisToDB(
	ctx context.Context,
	startDate time.Time,
	endDate time.Time,
) error {
	log.Printf("Dumping procurements from %s to %s from vmedis to DB", startDate, endDate)

	procurements, err := s.vmedis.GetAllProcurementsBetweenDates(ctx, startDate, endDate)
	if err != nil {
		return fmt.Errorf("get all procurements between %s and %s from vmedis: %w", startDate, endDate, err)
	}

	log.Printf("Got %d procurements from vmedis", len(procurements))

	if err := s.db.UpsertVmedisProcurements(ctx, procurements); err != nil {
		return fmt.Errorf("upsert vmedis procurements: %w", err)
	}

	log.Printf("Dumped %d procurements from vmedis to DB", len(procurements))

	var updatedDrugs []*kafkapb.UpdatedDrugByVmedisCode
	for _, p := range procurements {
		for _, pu := range p.ProcurementUnits {
			updatedDrugs = append(updatedDrugs, &kafkapb.UpdatedDrugByVmedisCode{
				RequestKey: fmt.Sprintf("procurement:%s:%s", p.InvoiceNumber, pu.DrugCode),
				VmedisCode: pu.DrugCode,
			})
		}
	}

	if err := s.drugProducer.ProduceUpdatedDrugByVmedisCode(ctx, updatedDrugs); err != nil {
		return fmt.Errorf("produce updated drug by vmedis code: %w", err)
	}

	log.Printf("Produced %d updated drugs by vmedis code", len(updatedDrugs))

	return nil
}

func NewService(
	db *gorm.DB,
	vmedisClient *vmedis.Client,
	drugProducer UpdatedDrugProducer,
) *Service {
	return &Service{
		db:           NewDatabase(db),
		vmedis:       vmedisClient,
		drugProducer: drugProducer,
	}
}
