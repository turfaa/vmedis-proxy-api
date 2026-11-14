package purchaseorder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

var (
	// ErrNotConfigured is returned when the pharmacy details printed on
	// purchase orders are not configured yet.
	ErrNotConfigured = errors.New("surat pesanan belum dikonfigurasi")
)

// ValidationError is returned when a create request is invalid.
type ValidationError struct {
	Message string
	// InvalidItems are the indexes of the invalid items in the request.
	InvalidItems []int
}

func (e *ValidationError) Error() string {
	return e.Message
}

type Service struct {
	db     *Database
	config Config
	now    func() time.Time
}

func NewService(db *gorm.DB, config Config) *Service {
	return &Service{
		db:     NewDatabase(db),
		config: config,
		now:    time.Now,
	}
}

// GetOptions returns the templates, the signer and the vmedis supplier page,
// for the create dialog.
func (s *Service) GetOptions() (OptionsResponse, error) {
	if !s.config.Enabled() {
		return OptionsResponse{}, ErrNotConfigured
	}

	return OptionsResponse{
		Templates:             slices2.Map(templates, Template.Info),
		Signer:                s.config.signer(),
		VmedisSupplierPageURL: s.config.VmedisSupplierPageURL(),
	}, nil
}

// Create numbers and stores a purchase order. createdByEmail is the login
// account making the request.
func (s *Service) Create(ctx context.Context, req CreateRequest, createdByEmail string) (CreateResponse, error) {
	if !s.config.Enabled() {
		return CreateResponse{}, ErrNotConfigured
	}

	if err := validateCreateRequest(req); err != nil {
		return CreateResponse{}, err
	}

	template, ok := templateByID(req.TemplateID)
	if !ok {
		return CreateResponse{}, fmt.Errorf("template %q: %w", req.TemplateID, ErrNotFound)
	}

	supplier, err := s.db.GetSupplier(ctx, req.SupplierVmedisID)
	if err != nil {
		return CreateResponse{}, err
	}

	var signer *models.PurchaseOrderSigner
	if req.IncludePharmacistSignature {
		signer = &models.PurchaseOrderSigner{
			Name:          s.config.PharmacistName,
			LicenseNumber: s.config.PharmacistLicenseNumber,
		}
	}

	items := make([]models.PurchaseOrderItem, len(req.Items))
	for i, item := range req.Items {
		var code *string
		if item.DrugVmedisCode != nil && strings.TrimSpace(*item.DrugVmedisCode) != "" {
			c := strings.TrimSpace(*item.DrugVmedisCode)
			code = &c
		}

		items[i] = models.PurchaseOrderItem{
			Position:       i + 1,
			DrugVmedisCode: code,
			DrugName:       strings.TrimSpace(item.DrugName),
			Quantity:       strings.TrimSpace(item.Quantity),
		}
	}

	// Truncated to microseconds, the precision postgres stores, so the stored
	// time is the one numbered here.
	createdAt := s.now().In(time.Local).Truncate(time.Microsecond)

	po := models.PurchaseOrder{
		CreatedAt:    createdAt,
		TemplateCode: template.ID,
		Year:         createdAt.Year(),
		Month:        int(createdAt.Month()),

		SupplierVmedisID: supplier.VmedisID,
		SupplierSnapshot: datatypes.NewJSONType(models.PurchaseOrderSupplier{
			VmedisID: supplier.VmedisID,
			Name:     supplier.Name,
			Address:  supplier.Address,
			City:     supplier.City,
			Phone:    supplier.Phone,
		}),

		IncludePharmacistSignature: req.IncludePharmacistSignature,
		SignerSnapshot:             datatypes.NewJSONType(signer),
		LetterheadSnapshot: datatypes.NewJSONType(models.PurchaseOrderLetterhead{
			Name:          s.config.PharmacyName,
			LicenseNumber: s.config.PharmacyLicenseNumber,
			Address:       s.config.PharmacyAddress,
			Email:         s.config.PharmacyEmail,
			Phone:         s.config.PharmacyPhone,
			City:          s.config.PharmacyCity,
		}),

		CreatedByEmployeeID:   req.CreatedByEmployee.ID,
		CreatedByEmployeeName: strings.TrimSpace(req.CreatedByEmployee.Name),
		CreatedByEmail:        createdByEmail,

		Items: items,
	}

	if err := s.db.Create(ctx, &po, func(sequence int) string {
		return formatNumber(template.NumberCode, po.Year, po.Month, sequence)
	}); err != nil {
		return CreateResponse{}, err
	}

	return CreateResponse{ID: po.ID, Number: po.Number, CreatedAt: po.CreatedAt}, nil
}

func validateCreateRequest(req CreateRequest) error {
	if req.SupplierVmedisID <= 0 {
		return &ValidationError{Message: "supplier wajib dipilih"}
	}

	if req.CreatedByEmployee.ID <= 0 || strings.TrimSpace(req.CreatedByEmployee.Name) == "" {
		return &ValidationError{Message: "pembuat surat pesanan wajib dipilih"}
	}

	if len(req.Items) == 0 {
		return &ValidationError{Message: "surat pesanan harus berisi minimal satu obat"}
	}

	var invalid []int
	for i, item := range req.Items {
		if strings.TrimSpace(item.DrugName) == "" || strings.TrimSpace(item.Quantity) == "" {
			invalid = append(invalid, i)
		}
	}

	if len(invalid) > 0 {
		return &ValidationError{
			Message:      fmt.Sprintf("%d obat belum diisi nama atau jumlahnya", len(invalid)),
			InvalidItems: invalid,
		}
	}

	return nil
}

// List returns a page of the purchase orders matching the filter, newest
// first.
func (s *Service) List(ctx context.Context, filter ListFilter) (ListResponse, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}

	if filter.PageSize < 1 {
		filter.PageSize = DefaultPageSize
	}

	filter.PageSize = min(filter.PageSize, MaxPageSize)

	pos, itemCounts, total, err := s.db.List(ctx, filter)
	if err != nil {
		return ListResponse{}, err
	}

	summaries := make([]Summary, len(pos))
	for i, po := range pos {
		summaries[i] = summaryFromDB(po, itemCounts[po.ID])
	}

	return ListResponse{
		PurchaseOrders: summaries,
		Total:          total,
		Page:           filter.Page,
		PageSize:       filter.PageSize,
	}, nil
}

// ListCreators returns the employees who have created at least one purchase
// order.
func (s *Service) ListCreators(ctx context.Context) (CreatorsResponse, error) {
	creators, err := s.db.ListCreators(ctx)
	if err != nil {
		return CreatorsResponse{}, err
	}

	if creators == nil {
		creators = []Employee{}
	}

	return CreatorsResponse{Creators: creators}, nil
}

// Get returns the purchase order with its items.
func (s *Service) Get(ctx context.Context, id uint) (Detail, error) {
	po, err := s.db.Get(ctx, id)
	if err != nil {
		return Detail{}, err
	}

	return detailFromDB(po), nil
}

// RenderedPDF is a rendered purchase order.
type RenderedPDF struct {
	Filename string
	Content  []byte
}

// RenderPDF renders the purchase order from its stored snapshots, so it is
// the same document every time.
func (s *Service) RenderPDF(ctx context.Context, id uint) (RenderedPDF, error) {
	po, err := s.db.Get(ctx, id)
	if err != nil {
		return RenderedPDF{}, err
	}

	template, ok := templateByID(po.TemplateCode)
	if !ok {
		return RenderedPDF{}, fmt.Errorf("purchase order %d has unknown template %q", po.ID, po.TemplateCode)
	}

	var buf bytes.Buffer
	if err := template.Render(&buf, documentFromDB(po)); err != nil {
		return RenderedPDF{}, fmt.Errorf("render purchase order %d: %w", po.ID, err)
	}

	return RenderedPDF{
		Filename: pdfFilename(po.CreatedAt, po.SupplierSnapshot.Data().Name, template.Name, po.Number),
		Content:  buf.Bytes(),
	}, nil
}
