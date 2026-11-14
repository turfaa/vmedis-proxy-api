package purchaseorder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/database"
	"github.com/turfaa/vmedis-proxy-api/database/models"
)

var jakarta = mustLoadLocation("Asia/Jakarta")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

// TestMain runs the tests in Asia/Jakarta, as main does for the server.
func TestMain(m *testing.M) {
	time.Local = jakarta
	m.Run()
}

func testConfig() Config {
	return Config{
		PharmacyName:            "Apotek Aulia Farma",
		PharmacyLicenseNumber:   "06012200432370001",
		PharmacyAddress:         "Jalan Pamagersari No. 10, Tanjungsari, Kabupaten Sumedang",
		PharmacyEmail:           "kontak@example.com",
		PharmacyPhone:           "0800000000",
		PharmacyCity:            "Sumedang",
		PharmacistName:          "Nama Apoteker",
		PharmacistLicenseNumber: "SIPA-123",
		VmedisBaseURL:           "https://aulia.vmedis.com/",
	}
}

type testEnv struct {
	db      *gorm.DB
	service *Service
	now     time.Time
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	db := openTestDB(t)

	suppliers := []models.Supplier{
		{VmedisID: 1, Name: "PT Supplier Satu", Address: "Jl. Satu", City: "Bandung", Phone: "022", Active: true},
		{VmedisID: 2, Name: "CV Dua: Farma", Active: true},
	}
	if err := db.Create(&suppliers).Error; err != nil {
		t.Fatalf("create suppliers: %v", err)
	}

	env := &testEnv{
		db:      db,
		service: NewService(db, testConfig()),
		now:     time.Date(2026, time.November, 14, 10, 0, 0, 0, jakarta),
	}
	env.service.now = func() time.Time { return env.now }

	return env
}

// openTestDB opens a fresh database: SQLite, or a new schema in the postgres
// database of PURCHASE_ORDER_TEST_POSTGRES_DSN when it is set.
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("PURCHASE_ORDER_TEST_POSTGRES_DSN")
	if dsn == "" {
		// The busy timeout lets concurrent test writers wait for each other
		// instead of failing with "database is locked".
		db, err := database.SqliteDB(t.TempDir() + "/test.db?_pragma=busy_timeout(10000)")
		if err != nil {
			t.Fatalf("open sqlite database: %v", err)
		}
		return db
	}

	schema := fmt.Sprintf("test_%d", time.Now().UnixNano())
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	if err := admin.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}
	t.Cleanup(func() { admin.Exec("DROP SCHEMA " + schema + " CASCADE") })

	db, err := database.PostgresDB(dsn + " search_path=" + schema)
	if err != nil {
		t.Fatalf("open postgres database: %v", err)
	}
	return db
}

func createRequest(supplierID int, items ...CreateItem) CreateRequest {
	if len(items) == 0 {
		code := "OBT1"
		items = []CreateItem{
			{DrugVmedisCode: &code, DrugName: "Paracetamol 500 mg", Quantity: "10 Box"},
			{DrugName: "Obat Tambahan", Quantity: "2 Strip"},
		}
	}

	return CreateRequest{
		TemplateID:                 "reguler",
		SupplierVmedisID:           supplierID,
		IncludePharmacistSignature: true,
		CreatedByEmployee:          Employee{ID: 7, Name: "Karyawan"},
		Items:                      items,
	}
}

func (e *testEnv) create(t *testing.T, req CreateRequest) CreateResponse {
	t.Helper()

	res, err := e.service.Create(context.Background(), req, "login@example.com")
	if err != nil {
		t.Fatalf("create purchase order: %v", err)
	}
	return res
}

func TestCreateNumbersPerTemplateAndMonth(t *testing.T) {
	env := newTestEnv(t)

	if got := env.create(t, createRequest(1)).Number; got != "001/SP-REG/AAF/XI/2026" {
		t.Errorf("first number = %q, want 001/SP-REG/AAF/XI/2026", got)
	}
	if got := env.create(t, createRequest(2)).Number; got != "002/SP-REG/AAF/XI/2026" {
		t.Errorf("second number = %q, want 002/SP-REG/AAF/XI/2026", got)
	}

	// 1 December 00:30 in Jakarta is still 30 November in UTC; the series
	// follows Jakarta.
	env.now = time.Date(2026, time.December, 1, 0, 30, 0, 0, jakarta)
	if got := env.create(t, createRequest(1)).Number; got != "001/SP-REG/AAF/XII/2026" {
		t.Errorf("first number in December = %q, want 001/SP-REG/AAF/XII/2026", got)
	}
}

func TestCreateConcurrentNumbersAreUnique(t *testing.T) {
	env := newTestEnv(t)

	const n = 20
	numbers := make([]string, n)
	errs := make([]error, n)

	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			res, err := env.service.Create(context.Background(), createRequest(1), "login@example.com")
			numbers[i], errs[i] = res.Number, err
		})
	}
	wg.Wait()

	seen := map[string]bool{}
	for i := range n {
		if errs[i] != nil {
			t.Fatalf("create %d: %v", i, errs[i])
		}
		if seen[numbers[i]] {
			t.Errorf("number %s given twice", numbers[i])
		}
		seen[numbers[i]] = true
	}

	for seq := 1; seq <= n; seq++ {
		if want := fmt.Sprintf("%03d/SP-REG/AAF/XI/2026", seq); !seen[want] {
			t.Errorf("number %s was skipped", want)
		}
	}
}

func TestCreateValidation(t *testing.T) {
	env := newTestEnv(t)

	tests := []struct {
		name         string
		req          CreateRequest
		wantInvalid  []int
		wantNotFound bool
	}{
		{name: "no supplier", req: createRequest(0)},
		{name: "no creator", req: func() CreateRequest { r := createRequest(1); r.CreatedByEmployee = Employee{}; return r }()},
		{name: "no items", req: func() CreateRequest { r := createRequest(1); r.Items = []CreateItem{}; return r }()},
		{
			name: "empty name or quantity",
			req: createRequest(1,
				CreateItem{DrugName: "A", Quantity: "1 Box"},
				CreateItem{DrugName: "B", Quantity: "  "},
				CreateItem{DrugName: "", Quantity: "1 Box"},
			),
			wantInvalid: []int{1, 2},
		},
		{name: "unknown supplier", req: createRequest(99), wantNotFound: true},
		{name: "unknown template", req: func() CreateRequest { r := createRequest(1); r.TemplateID = "x"; return r }()},
		{
			name: "too long",
			req: createRequest(1,
				CreateItem{DrugName: strings.Repeat("a", MaxDrugNameLength), Quantity: strings.Repeat("1", MaxQuantityLength)},
				CreateItem{DrugName: strings.Repeat("a", MaxDrugNameLength+1), Quantity: "1"},
				CreateItem{DrugName: "A", Quantity: strings.Repeat("1", MaxQuantityLength+1)},
			),
			wantInvalid: []int{1, 2},
		},
		{name: "too many items", req: createRequest(1, make([]CreateItem, MaxItems+1)...)},
		{name: "too long idempotency key", req: func() CreateRequest {
			r := createRequest(1)
			r.IdempotencyKey = strings.Repeat("k", MaxIdempotencyKeyLen+1)
			return r
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := env.service.Create(context.Background(), tt.req, "login@example.com")

			if tt.wantNotFound {
				if !errors.Is(err, ErrNotFound) {
					t.Fatalf("got error %v, want ErrNotFound", err)
				}
				return
			}

			var vErr *ValidationError
			if !errors.As(err, &vErr) {
				t.Fatalf("got error %v, want a ValidationError", err)
			}
			if fmt.Sprint(vErr.InvalidItems) != fmt.Sprint(tt.wantInvalid) {
				t.Errorf("invalid items = %v, want %v", vErr.InvalidItems, tt.wantInvalid)
			}
		})
	}

	// Rejected requests use no number up.
	if got := env.create(t, createRequest(1)).Number; got != "001/SP-REG/AAF/XI/2026" {
		t.Errorf("number after rejected requests = %q, want 001/SP-REG/AAF/XI/2026", got)
	}
}

func TestCreateIdempotent(t *testing.T) {
	env := newTestEnv(t)

	req := createRequest(1)
	req.IdempotencyKey = "key-1"

	first := env.create(t, req)
	if first.Replayed {
		t.Errorf("first create replayed")
	}

	// A retry after a lost response, even with changed choices, returns the
	// first purchase order.
	req.SupplierVmedisID = 2
	again := env.create(t, req)
	if !again.Replayed || again.ID != first.ID || again.Number != first.Number {
		t.Errorf("retry = %+v, want a replay of %+v", again, first)
	}

	// The replay used no number up.
	if got := env.create(t, createRequest(1)).Number; got != "002/SP-REG/AAF/XI/2026" {
		t.Errorf("next number = %q, want 002/SP-REG/AAF/XI/2026", got)
	}
}

func TestCreateIdempotentConcurrent(t *testing.T) {
	env := newTestEnv(t)

	const n = 10
	results := make([]CreateResponse, n)
	errs := make([]error, n)

	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			req := createRequest(1)
			req.IdempotencyKey = "same-key"
			results[i], errs[i] = env.service.Create(context.Background(), req, "")
		})
	}
	wg.Wait()

	for i := range n {
		if errs[i] != nil {
			t.Fatalf("create %d: %v", i, errs[i])
		}
		if results[i].ID != results[0].ID {
			t.Errorf("create %d got purchase order %d, want %d", i, results[i].ID, results[0].ID)
		}
	}

	if got := env.create(t, createRequest(1)).Number; got != "002/SP-REG/AAF/XI/2026" {
		t.Errorf("next number = %q, want 002/SP-REG/AAF/XI/2026: retries used numbers up", got)
	}
}

func TestFailedCreateUsesNoNumber(t *testing.T) {
	env := newTestEnv(t)

	// A row that holds the next number in another series makes the insert
	// fail on the unique number.
	blocker := models.PurchaseOrder{
		CreatedAt: env.now, TemplateCode: "other", Year: 2026, Month: 11, Sequence: 1,
		Number: "001/SP-REG/AAF/XI/2026", CreatedByEmployeeName: "x",
	}
	if err := env.db.Create(&blocker).Error; err != nil {
		t.Fatalf("create blocker: %v", err)
	}

	if _, err := env.service.Create(context.Background(), createRequest(1), ""); err == nil {
		t.Fatalf("create succeeded, want it to fail on the duplicate number")
	}

	if err := env.db.Delete(&blocker).Error; err != nil {
		t.Fatalf("delete blocker: %v", err)
	}

	if got := env.create(t, createRequest(1)).Number; got != "001/SP-REG/AAF/XI/2026" {
		t.Errorf("number after a failed create = %q, want 001/SP-REG/AAF/XI/2026", got)
	}
}

func TestNotConfigured(t *testing.T) {
	env := newTestEnv(t)
	env.service.config = Config{VmedisBaseURL: "https://aulia.vmedis.com"}

	if _, err := env.service.GetOptions(); !errors.Is(err, ErrNotConfigured) {
		t.Errorf("GetOptions error = %v, want ErrNotConfigured", err)
	}
	if _, err := env.service.Create(context.Background(), createRequest(1), ""); !errors.Is(err, ErrNotConfigured) {
		t.Errorf("Create error = %v, want ErrNotConfigured", err)
	}
}

func TestConfigValidate(t *testing.T) {
	if err := (Config{}).Validate(); err != nil {
		t.Errorf("empty config: %v, want nil (disabled)", err)
	}
	if err := testConfig().Validate(); err != nil {
		t.Errorf("full config: %v, want nil", err)
	}

	partial := testConfig()
	partial.PharmacistLicenseNumber = ""
	if err := partial.Validate(); err == nil || !strings.Contains(err.Error(), "pharmacist_license_number") {
		t.Errorf("partial config: %v, want an error naming pharmacist_license_number", err)
	}
}

func TestGetOptions(t *testing.T) {
	env := newTestEnv(t)

	options, err := env.service.GetOptions()
	if err != nil {
		t.Fatalf("get options: %v", err)
	}

	if options.VmedisSupplierPageURL != "https://aulia.vmedis.com/apt-data-supplier-v2" {
		t.Errorf("vmedis supplier page = %q", options.VmedisSupplierPageURL)
	}
	if len(options.Templates) != 1 || options.Templates[0].ID != "reguler" || options.Templates[0].Name != "Surat Pesanan Reguler" {
		t.Errorf("templates = %+v, want only reguler", options.Templates)
	}
	if options.Signer.Name != "Nama Apoteker" || options.Signer.LicenseNumber != "SIPA-123" {
		t.Errorf("signer = %+v", options.Signer)
	}
}

func TestGetKeepsSnapshots(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	unsigned := createRequest(1)
	unsigned.IncludePharmacistSignature = false
	created := env.create(t, unsigned)

	// Changes after the purchase order was made don't show on it.
	if err := env.db.Model(&models.Supplier{}).Where("vmedis_id = 1").Update("address", "Alamat Baru").Error; err != nil {
		t.Fatalf("update supplier: %v", err)
	}

	detail, err := env.service.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if detail.Supplier.Address != "Jl. Satu" || detail.Supplier.City != "Bandung" {
		t.Errorf("supplier = %+v, want the snapshot", detail.Supplier)
	}
	if detail.Signer != nil || detail.IncludePharmacistSignature {
		t.Errorf("signer = %+v, want none", detail.Signer)
	}
	if detail.CreatedByEmployee != (Employee{ID: 7, Name: "Karyawan"}) || detail.CreatedByEmail != "login@example.com" {
		t.Errorf("creator = %+v %q", detail.CreatedByEmployee, detail.CreatedByEmail)
	}
	if len(detail.Items) != 2 || detail.Items[0].Position != 1 || *detail.Items[0].DrugVmedisCode != "OBT1" || detail.Items[1].DrugVmedisCode != nil {
		t.Errorf("items = %+v", detail.Items)
	}
	if !detail.CreatedAt.Equal(env.now) {
		t.Errorf("created at = %v, want %v", detail.CreatedAt, env.now)
	}

	if _, err := env.service.Get(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("get unknown: %v, want ErrNotFound", err)
	}
}

func TestListFilters(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	env.now = time.Date(2026, time.October, 31, 23, 0, 0, 0, jakarta)
	october := env.create(t, createRequest(1, CreateItem{DrugName: "Amoxicillin", Quantity: "1 Box"}))

	env.now = time.Date(2026, time.November, 1, 8, 0, 0, 0, jakarta)
	byOther := createRequest(2, CreateItem{DrugName: "Promag 100%", Quantity: "3 Box"})
	byOther.CreatedByEmployee = Employee{ID: 8, Name: "Lainnya"}
	november := env.create(t, byOther)

	env.now = time.Date(2026, time.November, 2, 8, 0, 0, 0, jakarta)
	latest := env.create(t, createRequest(1))

	ids := func(filter ListFilter) []uint {
		t.Helper()
		res, err := env.service.List(ctx, filter)
		if err != nil {
			t.Fatalf("list %+v: %v", filter, err)
		}
		out := []uint{}
		for _, po := range res.PurchaseOrders {
			out = append(out, po.ID)
		}
		return out
	}

	nov1, _ := time.ParseInLocation(time.DateOnly, "2026-11-01", jakarta)

	tests := []struct {
		name   string
		filter ListFilter
		want   []uint
	}{
		{"all, newest first", ListFilter{}, []uint{latest.ID, november.ID, october.ID}},
		{"drug name", ListFilter{Query: "amoxi"}, []uint{october.ID}},
		{"number", ListFilter{Query: "/x/"}, []uint{october.ID}},
		{"like wildcards are literal", ListFilter{Query: "100%"}, []uint{november.ID}},
		{"from", ListFilter{From: nov1}, []uint{latest.ID, november.ID}},
		{"until", ListFilter{Until: nov1.Add(-time.Nanosecond)}, []uint{october.ID}},
		{"supplier", ListFilter{SupplierVmedisID: 2}, []uint{november.ID}},
		{"template", ListFilter{TemplateID: "reguler"}, []uint{latest.ID, november.ID, october.ID}},
		{"unknown template", ListFilter{TemplateID: "prekursor"}, []uint{}},
		{"creator", ListFilter{CreatedByEmployeeID: 7}, []uint{latest.ID, october.ID}},
		{"combined", ListFilter{CreatedByEmployeeID: 7, From: nov1}, []uint{latest.ID}},
		{"paged", ListFilter{Page: 2, PageSize: 2}, []uint{october.ID}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ids(tt.filter); fmt.Sprint(got) != fmt.Sprint(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}

	res, err := env.service.List(ctx, ListFilter{PageSize: 1000})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if res.Total != 3 || res.Page != 1 || res.PageSize != MaxPageSize {
		t.Errorf("total/page/pageSize = %d/%d/%d", res.Total, res.Page, res.PageSize)
	}
	if first := res.PurchaseOrders[0]; first.ItemCount != 2 || first.TemplateName != "Surat Pesanan Reguler" || first.Supplier.Name != "PT Supplier Satu" {
		t.Errorf("summary = %+v", first)
	}

	creators, err := env.service.ListCreators(ctx)
	if err != nil {
		t.Fatalf("list creators: %v", err)
	}
	if fmt.Sprint(creators.Creators) != fmt.Sprint([]Employee{{7, "Karyawan"}, {8, "Lainnya"}}) {
		t.Errorf("creators = %v", creators.Creators)
	}
}

func TestRenderPDF(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	var items []CreateItem
	for i := range 60 {
		items = append(items, CreateItem{
			DrugName: fmt.Sprintf("Obat %d dengan nama yang cukup panjang sehingga harus dibungkus menjadi beberapa baris — µg ✓ 💊", i+1),
			Quantity: "10 Box",
		})
	}
	items = append(items, CreateItem{DrugName: strings.Repeat("Sangatpanjangtanpaspasi", 8), Quantity: "1"})

	created := env.create(t, createRequest(2, items...))

	pdf, err := env.service.RenderPDF(ctx, created.ID)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if !bytes.HasPrefix(pdf.Content, []byte("%PDF-")) {
		t.Errorf("content is not a PDF")
	}
	if want := "2026-11-14 - CV Dua Farma - Surat Pesanan Reguler - 001-SP-REG-AAF-XI-2026.pdf"; pdf.Filename != want {
		t.Errorf("filename = %q, want %q", pdf.Filename, want)
	}
	if pages := bytes.Count(pdf.Content, []byte("/Type /Page\n")); pages < 3 {
		t.Errorf("rendered %d pages, want at least 3 for 61 long items", pages)
	}

	again, err := env.service.RenderPDF(ctx, created.ID)
	if err != nil {
		t.Fatalf("render again: %v", err)
	}
	if !bytes.Equal(pdf.Content, again.Content) {
		t.Errorf("rendering the same purchase order twice gave different PDFs")
	}
}

func renderTestDocument(items int) Document {
	doc := Document{
		Number:    "001/SP-REG/AAF/XI/2026",
		CreatedAt: time.Date(2026, time.November, 14, 10, 0, 0, 0, jakarta),
		Supplier:  models.PurchaseOrderSupplier{Name: "PT Supplier", Address: "Jl. Satu", City: "Bandung"},
		Letterhead: models.PurchaseOrderLetterhead{
			Name: "Apotek Aulia Farma", LicenseNumber: "1", Address: "Alamat", Email: "e", Phone: "p", City: "Sumedang",
		},
		Signer: &models.PurchaseOrderSigner{Name: "Apoteker", LicenseNumber: "SIPA"},
	}
	for i := range items {
		doc.Items = append(doc.Items, models.PurchaseOrderItem{DrugName: fmt.Sprintf("Obat %d", i+1), Quantity: "10 Box"})
	}
	return doc
}

// TestLastRowMovesWithTail checks that when the rows fill the first page and
// the tail (closing and signature) doesn't fit after them, the last row moves
// to the next page with the tail, and only the last row.
func TestLastRowMovesWithTail(t *testing.T) {
	// The most rows that fit on the first page with no tail.
	fitting := 0
	for n := 1; ; n++ {
		l := newLetter(renderTestDocument(n))
		l.writeItems(0)
		if l.pdf.PageNo() > 1 {
			break
		}
		fitting = n
	}

	const tail = 60.0
	l := newLetter(renderTestDocument(fitting))
	l.writeItems(tail)

	if got := l.pdf.PageNo(); got != 2 {
		t.Fatalf("after %d rows with a %vmm tail, on page %d, want 2", fitting, tail, got)
	}

	// Page 2 has the table header and one row.
	rowHeight := tableLineHeight + 2*tablePadding
	if got, want := l.pdf.GetY(), pageMarginTop+2*rowHeight+0.01; got > want {
		t.Errorf("page 2 ends at %.1fmm, want only the header and the last row (%.1fmm)", got, want)
	}
	if got := l.pdf.GetY() + tail; got > l.pageBottom() {
		t.Errorf("the tail ends at %.1fmm, past the page bottom %.1fmm", got, l.pageBottom())
	}
}

// TestLongestItemsFitOnPage checks that rows at the length limits stay well
// under a page tall, so the table never splits a row.
func TestLongestItemsFitOnPage(t *testing.T) {
	doc := renderTestDocument(0)
	doc.Items = []models.PurchaseOrderItem{{
		DrugName: strings.Repeat("W", MaxDrugNameLength),
		Quantity: strings.Repeat("W", MaxQuantityLength),
	}}

	l := newLetter(doc)
	l.pdf.SetFont(fontFamily, "", tableFontSize)
	nameWidth := l.contentWidth() - tableNoWidth - tableQtyWidth
	row := itemRow{
		nameLines: wrapText(l.pdf, doc.Items[0].DrugName, nameWidth-2*tablePadding),
		qtyLines:  wrapText(l.pdf, doc.Items[0].Quantity, tableQtyWidth-2*tablePadding),
	}

	if h, page := row.height(), l.pageBottom()-pageMarginTop; h > page/4 {
		t.Errorf("a row at the length limits is %.1fmm tall, more than a quarter of the %.1fmm page", h, page)
	}

	if err := renderReguler(&bytes.Buffer{}, doc); err != nil {
		t.Fatalf("render: %v", err)
	}
}

func TestRenderUnsigned(t *testing.T) {
	doc := renderTestDocument(3)
	doc.Signer = nil

	var buf bytes.Buffer
	if err := renderReguler(&buf, doc); err != nil {
		t.Fatalf("render: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Errorf("content is not a PDF")
	}
}

func BenchmarkRenderReguler50Items(b *testing.B) {
	doc := renderTestDocument(50)

	for b.Loop() {
		if err := renderReguler(&bytes.Buffer{}, doc); err != nil {
			b.Fatal(err)
		}
	}
}
