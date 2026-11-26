package purchaseorder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
	// maxPage keeps the list offset far from overflowing.
	maxPage = 100_000

	// Limits on create requests. They keep every item row well under a page
	// tall, so the PDF table never has to split a row.
	MaxItems              = 500
	MaxDrugNameLength     = 200
	MaxQuantityLength     = 50
	MaxEmployeeNameLength = 100
	MaxIdempotencyKeyLen  = 100

	// Limits on controlled items (Prekursor and OOT).
	MaxControlledQuantity    = 999_999
	MaxSmallestPackageLength = 100
	MaxNoteLength            = 100
	// MaxVmedisDetailLength caps the active ingredient and the dosage form
	// and strength taken from vmedis, so a row stays well under a page tall.
	MaxVmedisDetailLength = 250

	// MaxItemDetailsCodes is the most drugs the item details APIs take.
	MaxItemDetailsCodes = MaxItems
	// MaxRefreshCodes is the most drugs refreshed from vmedis at once.
	MaxRefreshCodes = 50

	MaxCancelReasonLength = 500
)

var (
	// ErrNotConfigured is returned when the pharmacy details printed on
	// purchase orders are not configured yet.
	ErrNotConfigured = errors.New("surat pesanan belum dikonfigurasi")

	// ErrAlreadyCancelled is returned when cancelling a purchase order that
	// is already cancelled.
	ErrAlreadyCancelled = errors.New("surat pesanan sudah dibatalkan")

	// ErrVmedis is returned when refreshing drugs from vmedis fails.
	ErrVmedis = errors.New("gagal memuat data obat dari vmedis")
)

// DrugRefresher fetches a drug from vmedis and stores it in the drugs table.
type DrugRefresher interface {
	DumpDrugDetailsFromVmedisToDBByVmedisID(ctx context.Context, vmedisID int64) error
}

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
	// drugs refreshes drugs from vmedis; nil disables refreshing.
	drugs DrugRefresher
	now   func() time.Time
}

func NewService(db *gorm.DB, config Config, drugs DrugRefresher) *Service {
	return &Service{
		db:     NewDatabase(db),
		config: config,
		drugs:  drugs,
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
		VmedisDrugPageURL:     s.config.VmedisDrugPageURL(),
	}, nil
}

// Create numbers and stores a purchase order. createdByEmail is the login
// account making the request.
//
// When the request has an idempotency key that was used before, it returns
// the purchase order made with that key, with Replayed set, instead of
// creating another one. So a retry after a lost response never uses up a
// second number.
func (s *Service) Create(ctx context.Context, req CreateRequest, createdByEmail string) (CreateResponse, error) {
	if !s.config.Enabled() {
		return CreateResponse{}, ErrNotConfigured
	}

	if err := validateCreateRequest(req); err != nil {
		return CreateResponse{}, err
	}

	template, ok := templateByID(req.TemplateID)
	if !ok {
		return CreateResponse{}, &ValidationError{Message: fmt.Sprintf("template %q tidak dikenal", req.TemplateID)}
	}

	key := strings.TrimSpace(req.IdempotencyKey)
	if key != "" {
		if res, ok, err := s.replay(ctx, key); err != nil || ok {
			return res, err
		}
	}

	supplier, err := s.db.GetSupplier(ctx, req.SupplierVmedisID)
	if err != nil {
		return CreateResponse{}, err
	}

	includeSignature := req.IncludePharmacistSignature || template.SignatureMode == SignatureModeWet

	var signer *models.PurchaseOrderSigner
	if includeSignature {
		signer = &models.PurchaseOrderSigner{
			Name:          s.config.PharmacistName,
			LicenseNumber: s.config.PharmacistLicenseNumber,
		}
	}

	items := make([]models.PurchaseOrderItem, len(req.Items))
	for i, item := range req.Items {
		var code *string
		if c := itemCode(item); c != "" {
			code = &c
		}

		items[i] = models.PurchaseOrderItem{
			Position:       i + 1,
			DrugVmedisCode: code,
			DrugName:       strings.TrimSpace(item.DrugName),
			Quantity:       strings.TrimSpace(item.Quantity),
		}
	}

	if template.ItemFields == ItemFieldsControlled {
		if err := s.fillControlledItems(ctx, req.Items, items); err != nil {
			return CreateResponse{}, err
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

		IncludePharmacistSignature: includeSignature,
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
	if key != "" {
		po.IdempotencyKey = &key
	}

	if err := s.db.Create(ctx, &po, func(sequence int) string {
		return formatNumber(template.NumberCode, po.Year, po.Month, sequence)
	}); err != nil {
		// Two requests with the same key can race past the check above. The
		// loser's insert fails on the key's unique index and its transaction,
		// sequence included, rolls back; it then returns the winner's.
		if key != "" {
			if res, ok, replayErr := s.replay(ctx, key); replayErr == nil && ok {
				return res, nil
			}
		}
		return CreateResponse{}, err
	}

	return CreateResponse{ID: po.ID, Number: po.Number, CreatedAt: po.CreatedAt}, nil
}

func itemCode(item CreateItem) string {
	if item.DrugVmedisCode == nil {
		return ""
	}

	return strings.TrimSpace(*item.DrugVmedisCode)
}

// fillControlledItems validates the request's items for a template with
// controlled item fields, and fills them in on items: the active ingredient
// and the dosage form and strength from the synced vmedis drugs, the
// smallest package, note and quantity from the request, and the quantity in
// words.
func (s *Service) fillControlledItems(ctx context.Context, reqItems []CreateItem, items []models.PurchaseOrderItem) error {
	codes := make([]string, 0, len(reqItems))
	for _, item := range reqItems {
		if code := itemCode(item); code != "" {
			codes = append(codes, code)
		}
	}

	drugs, err := s.db.GetDrugs(ctx, codes)
	if err != nil {
		return err
	}

	var notInVmedis, incomplete, invalid []int
	for i, reqItem := range reqItems {
		drug, ok := drugs[itemCode(reqItem)]
		if !ok {
			notInVmedis = append(notInVmedis, i)
			continue
		}

		activeIngredient, dosageFormStrength := drugActiveIngredient(drug), strings.TrimSpace(drug.KemasanBentukSediaan)
		if activeIngredient == "" || dosageFormStrength == "" {
			incomplete = append(incomplete, i)
			continue
		}

		quantity, quantityOK := parseControlledQuantity(reqItem.Quantity)
		smallestPackage := strings.TrimSpace(reqItem.SmallestPackage)
		note := strings.TrimSpace(deref(reqItem.Note))

		if !quantityOK ||
			smallestPackage == "" ||
			utf8.RuneCountInString(smallestPackage) > MaxSmallestPackageLength ||
			utf8.RuneCountInString(note) > MaxNoteLength ||
			utf8.RuneCountInString(activeIngredient) > MaxVmedisDetailLength ||
			utf8.RuneCountInString(dosageFormStrength) > MaxVmedisDetailLength {
			invalid = append(invalid, i)
			continue
		}

		words := terbilang(quantity)
		items[i].Quantity = strconv.FormatUint(quantity, 10)
		items[i].ActiveIngredient = &activeIngredient
		items[i].DosageFormStrength = &dosageFormStrength
		items[i].SmallestPackage = &smallestPackage
		items[i].QuantityWords = &words
		if note != "" {
			items[i].Note = &note
		}
	}

	switch {
	case len(notInVmedis) > 0:
		return &ValidationError{
			Message:      fmt.Sprintf("%d obat tidak ada di Data Obat Vmedis, padahal surat pesanan ini hanya untuk obat dari Vmedis", len(notInVmedis)),
			InvalidItems: notInVmedis,
		}

	case len(incomplete) > 0:
		return &ValidationError{
			Message:      fmt.Sprintf("%d obat belum diisi zat aktif atau bentuk & kekuatannya di Data Obat Vmedis", len(incomplete)),
			InvalidItems: incomplete,
		}

	case len(invalid) > 0:
		return &ValidationError{
			Message: fmt.Sprintf(
				"%d obat tidak valid: jumlah harus angka bulat 1–%d, kemasan terkecil wajib diisi (maksimal %d karakter), keterangan maksimal %d karakter, dan zat aktif serta bentuk & kekuatan di Vmedis maksimal %d karakter",
				len(invalid), MaxControlledQuantity, MaxSmallestPackageLength, MaxNoteLength, MaxVmedisDetailLength,
			),
			InvalidItems: invalid,
		}
	}

	return nil
}

// parseControlledQuantity parses a controlled item's quantity: a whole
// number from 1 to MaxControlledQuantity, digits only.
func parseControlledQuantity(s string) (uint64, bool) {
	s = strings.TrimSpace(s)
	if s == "" || strings.Trim(s, "0123456789") != "" {
		return 0, false
	}

	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil || n < 1 || n > MaxControlledQuantity {
		return 0, false
	}

	return n, true
}

// drugActiveIngredient is the drug's Zat Aktif Prekursor in vmedis, else
// its Kandungan.
func drugActiveIngredient(drug models.Drug) string {
	if v := strings.TrimSpace(drug.ZatAktifPrekursor); v != "" {
		return v
	}

	return strings.TrimSpace(drug.Kandungan)
}

// GetItemDetails returns what the synced vmedis drugs have on the drugs with
// the codes, for the controlled item fields of the template. It does not
// call vmedis.
func (s *Service) GetItemDetails(ctx context.Context, req ItemDetailsRequest) (ItemDetailsResponse, error) {
	template, codes, err := validateItemDetailsRequest(req, MaxItemDetailsCodes)
	if err != nil {
		return ItemDetailsResponse{}, err
	}

	drugs, err := s.db.GetDrugs(ctx, codes)
	if err != nil {
		return ItemDetailsResponse{}, err
	}

	items := make([]ItemDetails, len(codes))
	for i, code := range codes {
		items[i] = itemDetails(template, code, drugs)
	}

	return ItemDetailsResponse{Items: items}, nil
}

// RefreshItemDetails fetches the drugs with the codes from vmedis, stores
// them, and returns their item details. It is for after the operator has
// filled a drug's data in vmedis.
func (s *Service) RefreshItemDetails(ctx context.Context, req ItemDetailsRequest) (ItemDetailsResponse, error) {
	_, codes, err := validateItemDetailsRequest(req, MaxRefreshCodes)
	if err != nil {
		return ItemDetailsResponse{}, err
	}

	if s.drugs == nil {
		return ItemDetailsResponse{}, fmt.Errorf("%w: memuat ulang dari vmedis tidak tersedia", ErrVmedis)
	}

	drugs, err := s.db.GetDrugs(ctx, codes)
	if err != nil {
		return ItemDetailsResponse{}, err
	}

	var errs []error
	for _, code := range codes {
		drug, ok := drugs[code]
		if !ok {
			continue
		}

		if err := s.drugs.DumpDrugDetailsFromVmedisToDBByVmedisID(ctx, drug.VmedisID); err != nil {
			errs = append(errs, fmt.Errorf("refresh drug %s: %w", code, err))
		}
	}

	if len(errs) > 0 {
		return ItemDetailsResponse{}, fmt.Errorf("%w: %w", ErrVmedis, errors.Join(errs...))
	}

	return s.GetItemDetails(ctx, req)
}

func validateItemDetailsRequest(req ItemDetailsRequest, maxCodes int) (Template, []string, error) {
	template, ok := templateByID(req.TemplateID)
	if !ok {
		return Template{}, nil, &ValidationError{Message: fmt.Sprintf("template %q tidak dikenal", req.TemplateID)}
	}

	var codes []string
	for _, code := range req.Codes {
		if code = strings.TrimSpace(code); code != "" && !slices.Contains(codes, code) {
			codes = append(codes, code)
		}
	}

	if len(codes) > maxCodes {
		return Template{}, nil, &ValidationError{Message: fmt.Sprintf("maksimal %d obat", maxCodes)}
	}

	return template, codes, nil
}

func itemDetails(template Template, code string, drugs map[string]models.Drug) ItemDetails {
	details := ItemDetails{
		DrugVmedisCode:  code,
		Units:           []ItemUnit{},
		MissingInVmedis: []string{},
	}

	drug, ok := drugs[code]
	if ok {
		details.Found = true
		details.ActiveIngredient = nonEmpty(drugActiveIngredient(drug))
		details.DosageFormStrength = nonEmpty(drug.KemasanBentukSediaan)
		details.VmedisGolongan = nonEmpty(drug.Golongan)
		details.ClassMatches = golonganMatches(template, drug.Golongan)

		for _, unit := range drug.Units {
			details.Units = append(details.Units, ItemUnit{
				Unit:                   unit.Unit,
				ParentUnit:             unit.ParentUnit,
				ConversionToParentUnit: unit.ConversionToParentUnit,
			})
		}
	}

	if details.ActiveIngredient == nil {
		details.MissingInVmedis = append(details.MissingInVmedis, "activeIngredient")
	}
	if details.DosageFormStrength == nil {
		details.MissingInVmedis = append(details.MissingInVmedis, "dosageFormStrength")
	}

	return details
}

func nonEmpty(s string) *string {
	if s = strings.TrimSpace(s); s == "" {
		return nil
	}

	return &s
}

// Cancel marks the purchase order as cancelled (dibatalkan). It stays in the
// history with its number. cancelledByEmail is the login account making the
// request.
func (s *Service) Cancel(ctx context.Context, id uint, req CancelRequest, cancelledByEmail string) (Detail, error) {
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return Detail{}, &ValidationError{Message: "alasan pembatalan wajib diisi"}
	}
	if utf8.RuneCountInString(reason) > MaxCancelReasonLength {
		return Detail{}, &ValidationError{Message: fmt.Sprintf("alasan pembatalan maksimal %d karakter", MaxCancelReasonLength)}
	}

	name := strings.TrimSpace(req.CancelledByEmployee.Name)
	if req.CancelledByEmployee.ID <= 0 || name == "" {
		return Detail{}, &ValidationError{Message: "karyawan yang membatalkan wajib dipilih"}
	}
	if utf8.RuneCountInString(name) > MaxEmployeeNameLength {
		return Detail{}, &ValidationError{Message: fmt.Sprintf("nama karyawan maksimal %d karakter", MaxEmployeeNameLength)}
	}

	cancelledAt := s.now().In(time.Local).Truncate(time.Microsecond)
	if err := s.db.Cancel(ctx, id, cancellation{
		at:           cancelledAt,
		reason:       reason,
		employeeID:   req.CancelledByEmployee.ID,
		employeeName: name,
		email:        cancelledByEmail,
	}); err != nil {
		return Detail{}, err
	}

	return s.Get(ctx, id)
}

// replay returns the purchase order created with the idempotency key, if
// any.
func (s *Service) replay(ctx context.Context, key string) (CreateResponse, bool, error) {
	po, err := s.db.GetByIdempotencyKey(ctx, key)
	if errors.Is(err, ErrNotFound) {
		return CreateResponse{}, false, nil
	}
	if err != nil {
		return CreateResponse{}, false, err
	}

	return CreateResponse{ID: po.ID, Number: po.Number, CreatedAt: po.CreatedAt, Replayed: true}, true, nil
}

func validateCreateRequest(req CreateRequest) error {
	if utf8.RuneCountInString(req.IdempotencyKey) > MaxIdempotencyKeyLen {
		return &ValidationError{Message: fmt.Sprintf("idempotency key maksimal %d karakter", MaxIdempotencyKeyLen)}
	}

	if req.SupplierVmedisID <= 0 {
		return &ValidationError{Message: "supplier wajib dipilih"}
	}

	creatorName := strings.TrimSpace(req.CreatedByEmployee.Name)
	if req.CreatedByEmployee.ID <= 0 || creatorName == "" {
		return &ValidationError{Message: "pembuat surat pesanan wajib dipilih"}
	}
	if utf8.RuneCountInString(creatorName) > MaxEmployeeNameLength {
		return &ValidationError{Message: fmt.Sprintf("nama pembuat maksimal %d karakter", MaxEmployeeNameLength)}
	}

	if len(req.Items) == 0 {
		return &ValidationError{Message: "surat pesanan harus berisi minimal satu obat"}
	}
	if len(req.Items) > MaxItems {
		return &ValidationError{Message: fmt.Sprintf("surat pesanan maksimal berisi %d obat", MaxItems)}
	}

	var empty, tooLong []int
	for i, item := range req.Items {
		name, qty := strings.TrimSpace(item.DrugName), strings.TrimSpace(item.Quantity)

		switch {
		case name == "" || qty == "":
			empty = append(empty, i)
		case utf8.RuneCountInString(name) > MaxDrugNameLength || utf8.RuneCountInString(qty) > MaxQuantityLength:
			tooLong = append(tooLong, i)
		}
	}

	if len(empty) > 0 {
		return &ValidationError{
			Message:      fmt.Sprintf("%d obat belum diisi nama atau jumlahnya", len(empty)),
			InvalidItems: empty,
		}
	}

	if len(tooLong) > 0 {
		return &ValidationError{
			Message: fmt.Sprintf(
				"%d obat terlalu panjang: nama maksimal %d karakter, jumlah maksimal %d karakter",
				len(tooLong), MaxDrugNameLength, MaxQuantityLength,
			),
			InvalidItems: tooLong,
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
	filter.Page = min(filter.Page, maxPage)

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
		Filename: pdfFilename(po.CreatedAt, po.SupplierSnapshot.Data().Name, template.Name, po.Number, po.CancelledAt != nil),
		Content:  buf.Bytes(),
	}, nil
}
