package rejecteddrug

import (
	"context"
	"fmt"
	"time"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"

	"gorm.io/gorm"
)

type Service struct {
	db *Database
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: NewDatabase(db)}
}

func (s *Service) GetRejectedDrugs(ctx context.Context, filters ListFilters) ([]RejectedDrug, error) {
	rejectedDrugs, err := s.db.GetRejectedDrugs(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("get rejected drugs: %w", err)
	}

	return slices2.Map(rejectedDrugs, FromDBRejectedDrug), nil
}

func (s *Service) GetRejectedDrugByID(ctx context.Context, id uint) (RejectedDrug, error) {
	rejectedDrug, err := s.db.GetRejectedDrugByID(ctx, id)
	if err != nil {
		return RejectedDrug{}, fmt.Errorf("get rejected drug %d: %w", id, err)
	}

	return FromDBRejectedDrug(rejectedDrug), nil
}

func (s *Service) CreateRejectedDrug(ctx context.Context, request CreateRejectedDrugRequest, createdBy string) (RejectedDrug, error) {
	rejectedDrug, err := s.db.CreateRejectedDrug(ctx, models.RejectedDrug{
		DrugName:   request.DrugName,
		Resolution: models.RejectedDrugResolutionUnresolved,
		CreatedBy:  createdBy,
	})
	if err != nil {
		return RejectedDrug{}, fmt.Errorf("create rejected drug: %w", err)
	}

	return FromDBRejectedDrug(rejectedDrug), nil
}

func (s *Service) UpdateRejectedDrug(ctx context.Context, id uint, request UpdateRejectedDrugRequest, updatedBy string) (RejectedDrug, error) {
	rejectedDrug, err := s.db.GetRejectedDrugByID(ctx, id)
	if err != nil {
		return RejectedDrug{}, fmt.Errorf("get rejected drug %d: %w", id, err)
	}

	if request.DrugName != nil {
		rejectedDrug.DrugName = *request.DrugName
	}

	if request.ResolutionNotes != nil {
		rejectedDrug.ResolutionNotes = *request.ResolutionNotes
	}

	if request.Resolution != nil && *request.Resolution != rejectedDrug.Resolution {
		if !request.Resolution.Valid() {
			return RejectedDrug{}, fmt.Errorf("invalid resolution: %s", *request.Resolution)
		}

		rejectedDrug.Resolution = *request.Resolution

		if *request.Resolution == models.RejectedDrugResolutionUnresolved {
			rejectedDrug.ResolvedAt = nil
			rejectedDrug.ResolvedBy = ""
		} else {
			now := time.Now()
			rejectedDrug.ResolvedAt = &now
			rejectedDrug.ResolvedBy = updatedBy
		}
	}

	rejectedDrug, err = s.db.SaveRejectedDrug(ctx, rejectedDrug)
	if err != nil {
		return RejectedDrug{}, fmt.Errorf("save rejected drug %d: %w", id, err)
	}

	return FromDBRejectedDrug(rejectedDrug), nil
}

func (s *Service) DeleteRejectedDrug(ctx context.Context, id uint) error {
	if err := s.db.DeleteRejectedDrug(ctx, id); err != nil {
		return fmt.Errorf("delete rejected drug %d: %w", id, err)
	}

	return nil
}

func (s *Service) GetResolutions() []models.RejectedDrugResolution {
	return models.AllRejectedDrugResolutions()
}
