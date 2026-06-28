package shift

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
	"github.com/turfaa/vmedis-proxy-api/vmedis/v1"

	"gorm.io/gorm"
)

type Service struct {
	db           *Database
	redisDB      *RedisDatabase
	vmedisClient *vmedisv1.Client
}

func NewService(db *gorm.DB, redisClient redis.UniversalClient, vmedisClient *vmedisv1.Client) *Service {
	return &Service{
		db:           NewDatabase(db),
		redisDB:      NewRedisDatabase(redisClient),
		vmedisClient: vmedisClient,
	}
}

func (s *Service) GetShiftByCode(ctx context.Context, code string) (Shift, error) {
	dbShift, err := s.db.GetShiftByCode(ctx, code)
	if err != nil {
		return Shift{}, fmt.Errorf("get shift from db by code %s: %w", code, err)
	}

	return ShiftFromDB(dbShift), nil
}

func (s *Service) GetShiftByVmedisID(ctx context.Context, vmedisID int) (Shift, error) {
	dbShift, err := s.db.GetShiftByVmedisID(ctx, vmedisID)
	if err != nil {
		return Shift{}, fmt.Errorf("get shift from db by vmedis id %d: %w", vmedisID, err)
	}

	return ShiftFromDB(dbShift), nil
}

func (s *Service) GetShiftsBetween(ctx context.Context, from time.Time, to time.Time) ([]Shift, error) {
	dbShifts, err := s.db.GetShiftsBetween(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("get shifts from db between %s and %s: %w", from, to, err)
	}

	shifts := slices2.Map(dbShifts, ShiftFromDB)
	return shifts, nil
}

func (s *Service) DumpShiftsFromVmedisToDB(ctx context.Context, from time.Time, to time.Time) error {
	token, acquired, err := s.redisDB.AcquireDumpLock(ctx)
	if err != nil {
		return fmt.Errorf("acquire shift dump lock: %w", err)
	}

	if !acquired {
		log.Println("Shifts are already being dumped by another process, skipping")
		return nil
	}

	defer func() {
		if err := s.redisDB.ReleaseDumpLock(context.WithoutCancel(ctx), token); err != nil {
			log.Printf("Failed to release shift dump lock: %s", err)
		}
	}()

	vmedisShifts, err := s.vmedisClient.GetAllShiftsBetweenTimes(ctx, from, to)
	if err != nil {
		return fmt.Errorf("get all shifts from vmedis between %s and %s: %w", from, to, err)
	}

	log.Printf("Dumping %d shifts from vmedis to db", len(vmedisShifts))
	if err := s.db.UpsertVmedisShifts(ctx, vmedisShifts); err != nil {
		return fmt.Errorf("upsert vmedis shifts to db: %w", err)
	}
	log.Printf("Dumped %d shifts from vmedis to db", len(vmedisShifts))

	return nil
}

func (s *Service) GetShiftDumpStatus(ctx context.Context) (ShiftDumpStatusResponse, error) {
	locked, err := s.redisDB.IsDumpLocked(ctx)
	if err != nil {
		return ShiftDumpStatusResponse{}, fmt.Errorf("get shift dump status: %w", err)
	}

	status := ShiftDumpStatusIdle
	if locked {
		status = ShiftDumpStatusDumping
	}

	return ShiftDumpStatusResponse{Status: status}, nil
}
