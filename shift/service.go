package shift

import (
	"context"
	"fmt"
	"time"

	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
	"gorm.io/gorm"
)

type Service struct {
	db           *Database
	vmedisClient *vmedis.Client
}

func NewService(db *gorm.DB, vmedisClient *vmedis.Client) *Service {
	return &Service{db: NewDatabase(db), vmedisClient: vmedisClient}
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

func (s *Service) DumpShiftsFromVmedisToDB(ctx context.Context, from time.Time, to time.Time) ([]Shift, error) {
	vmedisShifts, err := s.vmedisClient.GetAllShiftsBetweenTimes(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("get all shifts from vmedis between %s and %s: %w", from, to, err)
	}

	if err := s.db.UpsertVmedisShifts(ctx, vmedisShifts); err != nil {
		return nil, fmt.Errorf("upsert vmedis shifts to db: %w", err)
	}

	return slices2.Map(vmedisShifts, ShiftFromVmedis), nil
}
