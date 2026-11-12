package supplier

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

// ErrDumpInProgress is returned when a supplier dump is requested while
// another one is still running.
var ErrDumpInProgress = errors.New("a supplier dump is already in progress")

type Service struct {
	db      *Database
	redisDB *RedisDatabase
	vmedis  *vmedisv2.Client
}

func NewService(db *gorm.DB, redisClient redis.UniversalClient, vmedisClient *vmedisv2.Client) *Service {
	return &Service{
		db:      NewDatabase(db),
		redisDB: NewRedisDatabase(redisClient),
		vmedis:  vmedisClient,
	}
}

// GetSuppliers returns all the suppliers in the DB, ordered by name.
func (s *Service) GetSuppliers(ctx context.Context) ([]Supplier, error) {
	dbSuppliers, err := s.db.GetSuppliers(ctx)
	if err != nil {
		return nil, fmt.Errorf("get suppliers from DB: %w", err)
	}

	return slices2.Map(dbSuppliers, FromDBSupplier), nil
}

// DumpSuppliersFromVmedisToDB upserts all the suppliers in vmedis to the DB,
// and soft-deletes the suppliers in the DB that are no longer in vmedis.
// It returns ErrDumpInProgress if another dump is running.
func (s *Service) DumpSuppliersFromVmedisToDB(ctx context.Context) error {
	release, err := s.acquireDumpLock(ctx)
	if err != nil {
		return err
	}
	defer release()

	ctx, cancel := context.WithTimeout(ctx, dumpLockTTL)
	defer cancel()

	return s.dumpSuppliersFromVmedisToDB(ctx)
}

// StartDumpSuppliersFromVmedisToDB starts dumping the suppliers from vmedis to
// the DB in the background, and returns once the dump has started.
// It returns ErrDumpInProgress if another dump is running.
func (s *Service) StartDumpSuppliersFromVmedisToDB(ctx context.Context) error {
	release, err := s.acquireDumpLock(ctx)
	if err != nil {
		return err
	}

	go func() {
		defer release()

		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), dumpLockTTL)
		defer cancel()

		if err := s.dumpSuppliersFromVmedisToDB(ctx); err != nil {
			log.Printf("Failed to dump suppliers from vmedis to DB: %s", err)
		}
	}()

	return nil
}

// GetDumpStatus returns whether a supplier dump is in progress.
func (s *Service) GetDumpStatus(ctx context.Context) (DumpStatusResponse, error) {
	locked, err := s.redisDB.IsDumpLocked(ctx)
	if err != nil {
		return DumpStatusResponse{}, fmt.Errorf("get supplier dump status: %w", err)
	}

	status := DumpStatusIdle
	if locked {
		status = DumpStatusDumping
	}

	return DumpStatusResponse{Status: status}, nil
}

func (s *Service) acquireDumpLock(ctx context.Context) (release func(), err error) {
	token, acquired, err := s.redisDB.AcquireDumpLock(ctx)
	if err != nil {
		return nil, err
	}

	if !acquired {
		return nil, ErrDumpInProgress
	}

	return func() {
		if err := s.redisDB.ReleaseDumpLock(context.WithoutCancel(ctx), token); err != nil {
			log.Printf("Failed to release supplier dump lock: %s", err)
		}
	}, nil
}

func (s *Service) dumpSuppliersFromVmedisToDB(ctx context.Context) error {
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
