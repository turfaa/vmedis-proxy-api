package supplier

import (
	"context"
	"errors"
	"fmt"
	"log"

	"gorm.io/gorm"

	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

type Service struct {
	db     *Database
	vmedis *vmedisv2.Client
}

func NewService(db *gorm.DB, vmedisClient *vmedisv2.Client) *Service {
	return &Service{
		db:     NewDatabase(db),
		vmedis: vmedisClient,
	}
}

// DumpSuppliersFromVmedisToDB upserts all the suppliers in vmedis to the DB,
// and soft-deletes the suppliers in the DB that are no longer in vmedis.
func (s *Service) DumpSuppliersFromVmedisToDB(ctx context.Context) error {
	log.Println("Dumping suppliers from vmedis to DB")

	suppliers, err := s.vmedis.GetAllSuppliers(ctx)
	if err != nil {
		return fmt.Errorf("get suppliers from vmedis: %w", err)
	}

	// An empty list is more likely a vmedis hiccup than every supplier being
	// deleted, and acting on it would soft-delete all of them.
	if len(suppliers) == 0 {
		return errors.New("vmedis returned no suppliers")
	}

	if err := s.db.UpsertVmedisSuppliers(ctx, suppliers); err != nil {
		return fmt.Errorf("upsert suppliers to DB: %w", err)
	}

	log.Printf("Upserted %d suppliers to DB", len(suppliers))

	vmedisIDs := make([]int, len(suppliers))
	for i, supplier := range suppliers {
		vmedisIDs[i] = supplier.ID
	}

	deleted, err := s.db.SoftDeleteSuppliersNotIn(ctx, vmedisIDs)
	if err != nil {
		return err
	}

	log.Printf("Dumped suppliers from vmedis to DB: %d suppliers in vmedis, %d soft-deleted", len(suppliers), deleted)
	return nil
}
