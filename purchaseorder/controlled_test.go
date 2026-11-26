package purchaseorder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

// seedControlledDrugs adds drugs as synced from vmedis: OOT1 and OOT2 are
// complete, OOT3 has no zat aktif or bentuk & kekuatan, and KERAS1 is
// complete but tagged Obat Keras.
func seedControlledDrugs(t *testing.T, env *testEnv) {
	t.Helper()

	drugs := []models.Drug{
		{VmedisID: 101, VmedisCode: "OOT1", Name: "Tramadol 50 mg", Golongan: "Obat-Obatan Tertentu", ZatAktifPrekursor: "Tramadol HCl", Kandungan: "ignored", KemasanBentukSediaan: "Kapsul 50 mg"},
		{VmedisID: 102, VmedisCode: "OOT2", Name: "Dekstrometorfan", Golongan: "DEPAN", Kandungan: "Dekstrometorfan HBr", KemasanBentukSediaan: "Tablet 15 mg"},
		{VmedisID: 103, VmedisCode: "OOT3", Name: "Belum Lengkap", Golongan: ""},
		{VmedisID: 104, VmedisCode: "KERAS1", Name: "Obat Keras", Golongan: "obat  keras", Kandungan: "Amoksisilin", KemasanBentukSediaan: "Kaplet 500 mg"},
	}
	if err := env.db.Create(&drugs).Error; err != nil {
		t.Fatalf("create drugs: %v", err)
	}

	units := []models.DrugUnit{
		{DrugVmedisCode: "OOT1", Unit: "Strip", UnitOrder: 0},
		{DrugVmedisCode: "OOT1", Unit: "Box", ParentUnit: "Strip", ConversionToParentUnit: 10, UnitOrder: 1},
	}
	if err := env.db.Create(&units).Error; err != nil {
		t.Fatalf("create units: %v", err)
	}
}

func controlledItem(code, quantity, smallestPackage string) CreateItem {
	return CreateItem{DrugVmedisCode: &code, DrugName: "Obat " + code, Quantity: quantity, SmallestPackage: smallestPackage}
}

func controlledRequest(templateID string, items ...CreateItem) CreateRequest {
	req := createRequest(1, items...)
	req.TemplateID = templateID
	req.IncludePharmacistSignature = false
	return req
}

func TestCreateControlled(t *testing.T) {
	env := newTestEnv(t)
	seedControlledDrugs(t, env)
	ctx := context.Background()

	note := "  Untuk resep dokter "
	second := controlledItem("OOT2", "012", "Strip @ 10 Tablet")
	second.Note = &note
	created := env.create(t, controlledRequest("oot", controlledItem("OOT1", " 3 ", "Box @ 10 Strip"), second))

	if created.Number != "001/SP-OOT/AAF/XI/2026" {
		t.Errorf("number = %q, want 001/SP-OOT/AAF/XI/2026", created.Number)
	}
	if got := env.create(t, controlledRequest("prekursor", controlledItem("OOT1", "1", "Box"))).Number; got != "001/SP-PRE/AAF/XI/2026" {
		t.Errorf("prekursor number = %q, want its own series", got)
	}

	detail, err := env.service.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if !detail.IncludePharmacistSignature || detail.Signer == nil {
		t.Errorf("wet-signature template stored no signer: %+v", detail)
	}
	if detail.TemplateItemFields != ItemFieldsControlled || detail.Status != StatusActive || detail.Cancellation != nil {
		t.Errorf("detail = %+v", detail)
	}

	str := func(s *string) string {
		if s == nil {
			return "<nil>"
		}
		return *s
	}
	got := make([]string, len(detail.Items))
	for i, item := range detail.Items {
		got[i] = strings.Join([]string{item.Quantity, str(item.ActiveIngredient), str(item.DosageFormStrength), str(item.SmallestPackage), str(item.QuantityWords), str(item.Note)}, "|")
	}
	want := []string{
		"3|Tramadol HCl|Kapsul 50 mg|Box @ 10 Strip|tiga|<nil>",
		"12|Dekstrometorfan HBr|Tablet 15 mg|Strip @ 10 Tablet|dua belas|Untuk resep dokter",
	}
	if !slices.Equal(got, want) {
		t.Errorf("items =\n%v\nwant\n%v", got, want)
	}

	// The snapshot stays when vmedis changes.
	if err := env.db.Model(&models.Drug{}).Where("vmedis_code = ?", "OOT1").Update("zat_aktif_prekursor", "Changed").Error; err != nil {
		t.Fatalf("update drug: %v", err)
	}
	again, _ := env.service.Get(ctx, created.ID)
	if str(again.Items[0].ActiveIngredient) != "Tramadol HCl" {
		t.Errorf("active ingredient changed with vmedis to %q", str(again.Items[0].ActiveIngredient))
	}
}

func TestCreateControlledValidation(t *testing.T) {
	env := newTestEnv(t)
	seedControlledDrugs(t, env)

	longNote := strings.Repeat("x", MaxNoteLength+1)
	withNote := controlledItem("OOT1", "1", "Box")
	withNote.Note = &longNote

	for name, tc := range map[string]struct {
		items   []CreateItem
		invalid []int
		message string
	}{
		"added by hand": {
			items:   []CreateItem{controlledItem("OOT1", "1", "Box"), {DrugName: "Tambahan", Quantity: "1", SmallestPackage: "Box"}},
			invalid: []int{1}, message: "tidak ada di Data Obat Vmedis",
		},
		"unknown code": {
			items:   []CreateItem{controlledItem("NOPE", "1", "Box")},
			invalid: []int{0}, message: "tidak ada di Data Obat Vmedis",
		},
		"incomplete in vmedis": {
			items:   []CreateItem{controlledItem("OOT1", "1", "Box"), controlledItem("OOT3", "1", "Box")},
			invalid: []int{1}, message: "belum diisi zat aktif",
		},
		"quantity with unit": {
			items:   []CreateItem{controlledItem("OOT1", "3 box", "Box")},
			invalid: []int{0}, message: "angka bulat",
		},
		"zero quantity": {
			items:   []CreateItem{controlledItem("OOT1", "0", "Box")},
			invalid: []int{0}, message: "angka bulat",
		},
		"quantity too big": {
			items:   []CreateItem{controlledItem("OOT1", fmt.Sprint(MaxControlledQuantity+1), "Box")},
			invalid: []int{0}, message: "angka bulat",
		},
		"no smallest package": {
			items:   []CreateItem{controlledItem("OOT1", "1", " ")},
			invalid: []int{0}, message: "kemasan terkecil wajib",
		},
		"long note": {
			items:   []CreateItem{withNote},
			invalid: []int{0}, message: "keterangan maksimal",
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := env.service.Create(context.Background(), controlledRequest("oot", tc.items...), "")

			var validationErr *ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("err = %v, want a validation error", err)
			}
			if !slices.Equal(validationErr.InvalidItems, tc.invalid) || !strings.Contains(validationErr.Message, tc.message) {
				t.Errorf("err = %q %v, want %q %v", validationErr.Message, validationErr.InvalidItems, tc.message, tc.invalid)
			}
		})
	}

	// A wrong golongan is only a warning, so it does not block.
	env.create(t, controlledRequest("oot", controlledItem("KERAS1", "1", "Box")))

	// Nothing was numbered by the rejected requests.
	if got := env.create(t, controlledRequest("oot", controlledItem("OOT1", "1", "Box"))).Number; got != "002/SP-OOT/AAF/XI/2026" {
		t.Errorf("number = %q, want 002/SP-OOT/AAF/XI/2026", got)
	}
}

func TestGetItemDetails(t *testing.T) {
	env := newTestEnv(t)
	seedControlledDrugs(t, env)

	res, err := env.service.GetItemDetails(context.Background(), ItemDetailsRequest{
		TemplateID: "oot",
		Codes:      []string{"OOT1", " OOT2", "OOT3", "KERAS1", "NOPE", "OOT1", ""},
	})
	if err != nil {
		t.Fatalf("get item details: %v", err)
	}

	str := func(s *string) string {
		if s == nil {
			return "<nil>"
		}
		return *s
	}
	matches := func(b *bool) string {
		if b == nil {
			return "<nil>"
		}
		return fmt.Sprint(*b)
	}

	var got []string
	for _, item := range res.Items {
		got = append(got, fmt.Sprintf("%s found=%v %s|%s|%s match=%s missing=%v units=%d",
			item.DrugVmedisCode, item.Found, str(item.ActiveIngredient), str(item.DosageFormStrength),
			str(item.VmedisGolongan), matches(item.ClassMatches), item.MissingInVmedis, len(item.Units)))
	}
	want := []string{
		"OOT1 found=true Tramadol HCl|Kapsul 50 mg|Obat-Obatan Tertentu match=true missing=[] units=2",
		"OOT2 found=true Dekstrometorfan HBr|Tablet 15 mg|DEPAN match=<nil> missing=[] units=0",
		"OOT3 found=true <nil>|<nil>|<nil> match=<nil> missing=[activeIngredient dosageFormStrength] units=0",
		"KERAS1 found=true Amoksisilin|Kaplet 500 mg|obat  keras match=false missing=[] units=0",
		"NOPE found=false <nil>|<nil>|<nil> match=<nil> missing=[activeIngredient dosageFormStrength] units=0",
	}
	if !slices.Equal(got, want) {
		t.Errorf("item details =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}

	if units := res.Items[0].Units; units[1] != (ItemUnit{Unit: "Box", ParentUnit: "Strip", ConversionToParentUnit: 10}) {
		t.Errorf("units = %+v", units)
	}

	reguler, _ := env.service.GetItemDetails(context.Background(), ItemDetailsRequest{TemplateID: "reguler", Codes: []string{"KERAS1"}})
	if reguler.Items[0].ClassMatches != nil {
		t.Errorf("reguler is for any drug, got class matches %v", *reguler.Items[0].ClassMatches)
	}

	if _, err := env.service.GetItemDetails(context.Background(), ItemDetailsRequest{TemplateID: "nope"}); err == nil {
		t.Errorf("unknown template: no error")
	}
}

func TestRefreshItemDetails(t *testing.T) {
	env := newTestEnv(t)
	seedControlledDrugs(t, env)
	ctx := context.Background()

	env.refresher.updates = map[int64]map[string]any{
		103: {"zat_aktif_prekursor": "Pseudoefedrin HCl", "kemasan_bentuk_sediaan": "Tablet 60 mg"},
	}

	res, err := env.service.RefreshItemDetails(ctx, ItemDetailsRequest{TemplateID: "prekursor", Codes: []string{"OOT3", "NOPE"}})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}

	if got := res.Items[0]; got.ActiveIngredient == nil || *got.ActiveIngredient != "Pseudoefedrin HCl" || len(got.MissingInVmedis) != 0 {
		t.Errorf("refreshed item = %+v", got)
	}
	if !slices.Equal(env.refresher.refreshed, []int64{103}) {
		t.Errorf("refreshed %v, want only the known drug 103", env.refresher.refreshed)
	}

	env.refresher.fail = true
	if _, err := env.service.RefreshItemDetails(ctx, ItemDetailsRequest{TemplateID: "prekursor", Codes: []string{"OOT3"}}); !errors.Is(err, ErrVmedis) {
		t.Errorf("failed refresh: err = %v, want ErrVmedis", err)
	}

	tooMany := make([]string, MaxRefreshCodes+1)
	for i := range tooMany {
		tooMany[i] = fmt.Sprint("C", i)
	}
	var validationErr *ValidationError
	if _, err := env.service.RefreshItemDetails(ctx, ItemDetailsRequest{TemplateID: "oot", Codes: tooMany}); !errors.As(err, &validationErr) {
		t.Errorf("too many codes: err = %v", err)
	}
}

func TestCancel(t *testing.T) {
	env := newTestEnv(t)
	seedControlledDrugs(t, env)
	ctx := context.Background()

	created := env.create(t, controlledRequest("oot", controlledItem("OOT1", "2", "Box")))
	other := env.create(t, createRequest(1))

	before, err := env.service.RenderPDF(ctx, created.ID)
	if err != nil {
		t.Fatalf("render before cancelling: %v", err)
	}

	for name, req := range map[string]CancelRequest{
		"no reason":      {Reason: "  ", CancelledByEmployee: Employee{ID: 3, Name: "A"}},
		"long reason":    {Reason: strings.Repeat("x", MaxCancelReasonLength+1), CancelledByEmployee: Employee{ID: 3, Name: "A"}},
		"no employee":    {Reason: "salah"},
		"employee no id": {Reason: "salah", CancelledByEmployee: Employee{Name: "A"}},
	} {
		var validationErr *ValidationError
		if _, err := env.service.Cancel(ctx, created.ID, req, ""); !errors.As(err, &validationErr) {
			t.Errorf("%s: err = %v, want a validation error", name, err)
		}
	}

	env.now = env.now.Add(48 * time.Hour)
	req := CancelRequest{Reason: " Salah supplier ", CancelledByEmployee: Employee{ID: 3, Name: " Budi "}}
	detail, err := env.service.Cancel(ctx, created.ID, req, "login@example.com")
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}

	want := Cancellation{
		CancelledAt:         env.now,
		Reason:              "Salah supplier",
		CancelledByEmployee: Employee{ID: 3, Name: "Budi"},
		CancelledByEmail:    "login@example.com",
	}
	if detail.Status != StatusCancelled || detail.Cancellation == nil ||
		!detail.Cancellation.CancelledAt.Equal(want.CancelledAt) ||
		detail.Cancellation.Reason != want.Reason ||
		detail.Cancellation.CancelledByEmployee != want.CancelledByEmployee ||
		detail.Cancellation.CancelledByEmail != want.CancelledByEmail {
		t.Errorf("cancelled detail = %+v %+v", detail, detail.Cancellation)
	}

	if _, err := env.service.Cancel(ctx, created.ID, req, ""); !errors.Is(err, ErrAlreadyCancelled) {
		t.Errorf("cancel twice: err = %v, want ErrAlreadyCancelled", err)
	}
	if _, err := env.service.Cancel(ctx, 999, req, ""); !errors.Is(err, ErrNotFound) {
		t.Errorf("cancel unknown: err = %v, want ErrNotFound", err)
	}

	for status, want := range map[Status][]uint{
		"":              {other.ID, created.ID},
		StatusActive:    {other.ID},
		StatusCancelled: {created.ID},
	} {
		res, err := env.service.List(ctx, ListFilter{Status: status})
		if err != nil {
			t.Fatalf("list %q: %v", status, err)
		}

		var got []uint
		for _, po := range res.PurchaseOrders {
			got = append(got, po.ID)
		}
		if !slices.Equal(got, want) {
			t.Errorf("list %q = %v, want %v", status, got, want)
		}
	}

	after, err := env.service.RenderPDF(ctx, created.ID)
	if err != nil {
		t.Fatalf("render after cancelling: %v", err)
	}
	if want := "2026-11-14 - PT Supplier Satu - Surat Pesanan Obat-Obat Tertentu - 001-SP-OOT-AAF-XI-2026 - DIBATALKAN.pdf"; after.Filename != want {
		t.Errorf("filename = %q, want %q", after.Filename, want)
	}
	if bytes.Equal(before.Content, after.Content) {
		t.Errorf("the cancelled PDF is the same as before cancelling")
	}

	again, _ := env.service.RenderPDF(ctx, created.ID)
	if !bytes.Equal(after.Content, again.Content) {
		t.Errorf("rendering the cancelled purchase order twice gave different PDFs")
	}

	// A cancelled purchase order keeps its number, and the next one gets a new one.
	if got := env.create(t, controlledRequest("oot", controlledItem("OOT1", "2", "Box"))).Number; got != "002/SP-OOT/AAF/XI/2026" {
		t.Errorf("next number = %q, want 002/SP-OOT/AAF/XI/2026", got)
	}
}

func controlledTestDocument(items int, note bool) Document {
	doc := renderTestDocument(0)
	doc.Number = "001/SP-PRE/AAF/XI/2026"
	for i := range items {
		s := func(v string) *string { return &v }
		item := models.PurchaseOrderItem{
			DrugName:           fmt.Sprintf("Obat Prekursor %d dengan nama panjang", i+1),
			Quantity:           "12",
			ActiveIngredient:   s("Pseudoefedrin HCl"),
			DosageFormStrength: s("Tablet salut selaput 60 mg"),
			SmallestPackage:    s("Box @ 10 Strip"),
			QuantityWords:      s("dua belas"),
		}
		if note && i == 0 {
			item.Note = s("Untuk stok bulan depan")
		}
		doc.Items = append(doc.Items, item)
	}
	return doc
}

func TestRenderControlled(t *testing.T) {
	for _, render := range []func(*bytes.Buffer, Document) error{
		func(b *bytes.Buffer, d Document) error { return renderPrekursor(b, d) },
		func(b *bytes.Buffer, d Document) error { return renderOOT(b, d) },
	} {
		var short, long, cancelled bytes.Buffer
		if err := render(&short, controlledTestDocument(2, true)); err != nil {
			t.Fatalf("render: %v", err)
		}
		if pages := bytes.Count(short.Bytes(), []byte("/Type /Page\n")); pages != 1 {
			t.Errorf("2 items rendered on %d pages, want 1", pages)
		}

		if err := render(&long, controlledTestDocument(40, false)); err != nil {
			t.Fatalf("render long: %v", err)
		}
		if pages := bytes.Count(long.Bytes(), []byte("/Type /Page\n")); pages < 2 {
			t.Errorf("40 items rendered on %d pages, want at least 2", pages)
		}

		doc := controlledTestDocument(2, false)
		doc.Signer = nil
		doc.Cancellation = &DocumentCancellation{At: doc.CreatedAt, EmployeeName: "Budi 💊", Reason: "Salah"}
		if err := render(&cancelled, doc); err != nil {
			t.Fatalf("render cancelled: %v", err)
		}
	}
}

func TestControlledTableKetColumn(t *testing.T) {
	without := controlledTable(controlledTestDocument(2, false).Items, "Zat Aktif")
	with := controlledTable(controlledTestDocument(2, true).Items, "Zat Aktif")

	if len(with.columns) != len(without.columns)+1 || with.columns[len(with.columns)-1].header != "Ket." {
		t.Errorf("columns with a note: %d, without: %d", len(with.columns), len(without.columns))
	}

	l := newLetter(controlledTestDocument(2, false))
	total := 0.0
	for _, w := range l.columnWidths(without) {
		total += w
	}
	if total < l.contentWidth()-0.01 || total > l.contentWidth()+0.01 {
		t.Errorf("columns are %.2fmm wide, want the content width %.2fmm", total, l.contentWidth())
	}
}

func TestHandlerControlledStatuses(t *testing.T) {
	env := newTestEnv(t)
	seedControlledDrugs(t, env)
	r := newTestRouter(env.service)

	if rec := doRequest(r, http.MethodGet, "/purchase-orders/item-details?templateId=oot&codes=OOT1,OOT3", nil); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"missingInVmedis":["activeIngredient","dosageFormStrength"]`) {
		t.Errorf("item details: %d %s", rec.Code, rec.Body)
	}

	env.refresher.fail = true
	if rec := doRequest(r, http.MethodPost, "/purchase-orders/item-details/refresh", ItemDetailsRequest{TemplateID: "oot", Codes: []string{"OOT1"}}); rec.Code != http.StatusBadGateway {
		t.Errorf("failed refresh: %d %s", rec.Code, rec.Body)
	}

	created := env.create(t, createRequest(1))
	path := fmt.Sprintf("/purchase-orders/%d/cancel", created.ID)
	req := CancelRequest{Reason: "salah", CancelledByEmployee: Employee{ID: 1, Name: "A"}}

	if rec := doRequest(r, http.MethodPost, path, CancelRequest{}); rec.Code != http.StatusBadRequest {
		t.Errorf("invalid cancel: %d %s", rec.Code, rec.Body)
	}
	if rec := doRequest(r, http.MethodPost, path, req); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"cancelled"`) {
		t.Errorf("cancel: %d %s", rec.Code, rec.Body)
	}
	if rec := doRequest(r, http.MethodPost, path, req); rec.Code != http.StatusConflict {
		t.Errorf("cancel twice: %d %s", rec.Code, rec.Body)
	}
	if rec := doRequest(r, http.MethodPost, "/purchase-orders/999/cancel", req); rec.Code != http.StatusNotFound {
		t.Errorf("cancel unknown: %d %s", rec.Code, rec.Body)
	}
	if rec := doRequest(r, http.MethodGet, "/purchase-orders?status=nope", nil); rec.Code != http.StatusBadRequest {
		t.Errorf("invalid status: %d %s", rec.Code, rec.Body)
	}
}
