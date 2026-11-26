package purchaseorder

import (
	"time"

	"github.com/turfaa/vmedis-proxy-api/database/models"
)

// OptionsResponse is the response of the purchase order options API: what
// the dashboard needs to show the create dialog.
type OptionsResponse struct {
	Templates []TemplateInfo `json:"templates"`
	Signer    Signer         `json:"signer"`
	// VmedisSupplierPageURL is the vmedis page where suppliers are edited.
	VmedisSupplierPageURL string `json:"vmedisSupplierPageUrl"`
	// VmedisDrugPageURL is the vmedis page where drugs are edited.
	VmedisDrugPageURL string `json:"vmedisDrugPageUrl"`
}

// TemplateInfo describes a purchase order template.
type TemplateInfo struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	ItemFields    ItemFields    `json:"itemFields"`
	SignatureMode SignatureMode `json:"signatureMode"`
}

// Signer is the pharmacist who signs purchase orders (Apoteker Penanggung
// Jawab).
type Signer struct {
	Name string `json:"name"`
	// LicenseNumber is the pharmacist's SIPA.
	LicenseNumber string `json:"licenseNumber"`
}

// Employee is an HRIS employee.
type Employee struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// CreateRequest is the request to create a purchase order.
type CreateRequest struct {
	// IdempotencyKey identifies the create attempt, e.g. a UUID the client
	// makes once per dialog opening. Retrying with the same key returns the
	// purchase order made by the first request instead of creating another.
	IdempotencyKey   string `json:"idempotencyKey"`
	TemplateID       string `json:"templateId"`
	SupplierVmedisID int    `json:"supplierVmedisId"`
	// IncludePharmacistSignature is ignored, and taken as true, on templates
	// with a wet signature.
	IncludePharmacistSignature bool         `json:"includePharmacistSignature"`
	CreatedByEmployee          Employee     `json:"createdByEmployee"`
	Items                      []CreateItem `json:"items"`
}

// CreateItem is an item of the purchase order to create.
type CreateItem struct {
	// DrugVmedisCode is null for drugs added by hand that are not in vmedis.
	DrugVmedisCode *string `json:"drugVmedisCode"`
	DrugName       string  `json:"drugName"`
	// Quantity is free text, e.g. "10 Box", on basic templates, and a whole
	// number, e.g. "12", on controlled ones.
	Quantity string `json:"quantity"`

	// SmallestPackage and Note are for templates with controlled item
	// fields, and ignored on the others. The active ingredient and the
	// dosage form and strength are not sent: they are taken from vmedis.
	SmallestPackage string  `json:"smallestPackage"`
	Note            *string `json:"note"`
}

// CreateResponse is the response of the create purchase order API.
type CreateResponse struct {
	ID        uint      `json:"id"`
	Number    string    `json:"number"`
	CreatedAt time.Time `json:"createdAt"`
	// Replayed is true when the request's idempotency key was used before,
	// and this is the purchase order created then.
	Replayed bool `json:"replayed"`
}

// ListFilter filters the purchase orders to list.
type ListFilter struct {
	// Query matches the number or any item's drug name, case-insensitively.
	Query string
	// From and Until bound the creation time, inclusive. Zero means unbounded.
	From  time.Time
	Until time.Time

	SupplierVmedisID    int
	TemplateID          string
	CreatedByEmployeeID int
	// Status is StatusActive, StatusCancelled, or empty for both.
	Status Status

	Page     int
	PageSize int
}

// ListResponse is the response of the list purchase orders API.
type ListResponse struct {
	PurchaseOrders []Summary `json:"purchaseOrders"`
	Total          int64     `json:"total"`
	Page           int       `json:"page"`
	PageSize       int       `json:"pageSize"`
}

// SummarySupplier is the supplier shown in a purchase order summary.
type SummarySupplier struct {
	VmedisID int    `json:"vmedisId"`
	Name     string `json:"name"`
}

// Status is whether a purchase order is cancelled.
type Status string

const (
	StatusActive    Status = "active"
	StatusCancelled Status = "cancelled"
)

// Cancellation is when, by whom and why a purchase order was cancelled.
type Cancellation struct {
	CancelledAt         time.Time `json:"cancelledAt"`
	Reason              string    `json:"reason"`
	CancelledByEmployee Employee  `json:"cancelledByEmployee"`
	CancelledByEmail    string    `json:"cancelledByEmail"`
}

// CancelRequest is the request to cancel a purchase order.
type CancelRequest struct {
	Reason              string   `json:"reason"`
	CancelledByEmployee Employee `json:"cancelledByEmployee"`
}

// ItemDetailsRequest asks for what vmedis has on the drugs to put on a
// purchase order of a template.
type ItemDetailsRequest struct {
	TemplateID string   `json:"templateId"`
	Codes      []string `json:"codes"`
}

// ItemDetailsResponse is the response of the item details APIs.
type ItemDetailsResponse struct {
	Items []ItemDetails `json:"items"`
}

// ItemDetails is what vmedis has on a drug, for the controlled item fields.
type ItemDetails struct {
	DrugVmedisCode string `json:"drugVmedisCode"`
	// Found is false when the drug is not in the synced vmedis drugs.
	Found bool `json:"found"`
	// ActiveIngredient is the drug's Zat Aktif Prekursor in vmedis, else its
	// Kandungan; nil when both are empty.
	ActiveIngredient *string `json:"activeIngredient"`
	// DosageFormStrength is the drug's Kemasan / Bentuk Sediaan in vmedis,
	// nil when empty.
	DosageFormStrength *string `json:"dosageFormStrength"`
	// Units are the drug's units, from the smallest, for the dashboard to
	// describe the smallest package from the unit ordered.
	Units []ItemUnit `json:"units"`
	// VmedisGolongan is the drug's golongan in vmedis, nil when empty.
	VmedisGolongan *string `json:"vmedisGolongan"`
	// ClassMatches is whether the golongan is the template's drug class:
	// nil when the golongan is not a drug class (e.g. a shelf) or the
	// template is for any drug.
	ClassMatches *bool `json:"classMatches"`
	// MissingInVmedis are the required fields that are empty in vmedis:
	// "activeIngredient" and "dosageFormStrength".
	MissingInVmedis []string `json:"missingInVmedis"`
}

// ItemUnit is a unit of a drug.
type ItemUnit struct {
	Unit string `json:"unit"`
	// ParentUnit is the next smaller unit, empty for the smallest one.
	ParentUnit string `json:"parentUnit"`
	// ConversionToParentUnit is how many of the parent unit make one of this
	// unit, 0 for the smallest unit.
	ConversionToParentUnit float64 `json:"conversionToParentUnit"`
}

// Summary is a purchase order in the list.
type Summary struct {
	ID                         uint            `json:"id"`
	Number                     string          `json:"number"`
	TemplateID                 string          `json:"templateId"`
	TemplateName               string          `json:"templateName"`
	TemplateItemFields         ItemFields      `json:"templateItemFields"`
	Status                     Status          `json:"status"`
	Supplier                   SummarySupplier `json:"supplier"`
	ItemCount                  int             `json:"itemCount"`
	IncludePharmacistSignature bool            `json:"includePharmacistSignature"`
	CreatedByEmployee          Employee        `json:"createdByEmployee"`
	CreatedByEmail             string          `json:"createdByEmail"`
	CreatedAt                  time.Time       `json:"createdAt"`
	Cancellation               *Cancellation   `json:"cancellation"`
}

// Supplier is the supplier snapshot of a purchase order.
type Supplier struct {
	VmedisID int    `json:"vmedisId"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	City     string `json:"city"`
	Phone    string `json:"phone"`
}

// Item is an item of a purchase order.
type Item struct {
	Position       int     `json:"position"`
	DrugVmedisCode *string `json:"drugVmedisCode"`
	DrugName       string  `json:"drugName"`
	Quantity       string  `json:"quantity"`
	// The fields below are set on templates with controlled item fields,
	// nil on the others.
	ActiveIngredient   *string `json:"activeIngredient"`
	DosageFormStrength *string `json:"dosageFormStrength"`
	SmallestPackage    *string `json:"smallestPackage"`
	QuantityWords      *string `json:"quantityWords"`
	Note               *string `json:"note"`
}

// Detail is a purchase order with its items.
type Detail struct {
	ID                         uint          `json:"id"`
	Number                     string        `json:"number"`
	TemplateID                 string        `json:"templateId"`
	TemplateName               string        `json:"templateName"`
	TemplateItemFields         ItemFields    `json:"templateItemFields"`
	Status                     Status        `json:"status"`
	Supplier                   Supplier      `json:"supplier"`
	ItemCount                  int           `json:"itemCount"`
	IncludePharmacistSignature bool          `json:"includePharmacistSignature"`
	Signer                     *Signer       `json:"signer"`
	CreatedByEmployee          Employee      `json:"createdByEmployee"`
	CreatedByEmail             string        `json:"createdByEmail"`
	CreatedAt                  time.Time     `json:"createdAt"`
	Cancellation               *Cancellation `json:"cancellation"`
	Items                      []Item        `json:"items"`
}

// CreatorsResponse is the response of the purchase order creators API.
type CreatorsResponse struct {
	Creators []Employee `json:"creators"`
}

func templateName(templateID string) string {
	if t, ok := templateByID(templateID); ok {
		return t.Name
	}

	return templateID
}

func templateItemFields(templateID string) ItemFields {
	if t, ok := templateByID(templateID); ok {
		return t.ItemFields
	}

	return ItemFieldsBasic
}

func statusFromDB(po models.PurchaseOrder) Status {
	if po.CancelledAt != nil {
		return StatusCancelled
	}

	return StatusActive
}

func cancellationFromDB(po models.PurchaseOrder) *Cancellation {
	if po.CancelledAt == nil {
		return nil
	}

	return &Cancellation{
		CancelledAt: *po.CancelledAt,
		Reason:      deref(po.CancelledReason),
		CancelledByEmployee: Employee{
			ID:   deref(po.CancelledByEmployeeID),
			Name: deref(po.CancelledByEmployeeName),
		},
		CancelledByEmail: deref(po.CancelledByEmail),
	}
}

func summaryFromDB(po models.PurchaseOrder, itemCount int) Summary {
	supplier := po.SupplierSnapshot.Data()

	return Summary{
		ID:                 po.ID,
		Number:             po.Number,
		TemplateID:         po.TemplateCode,
		TemplateName:       templateName(po.TemplateCode),
		TemplateItemFields: templateItemFields(po.TemplateCode),
		Status:             statusFromDB(po),
		Supplier: SummarySupplier{
			VmedisID: po.SupplierVmedisID,
			Name:     supplier.Name,
		},
		ItemCount:                  itemCount,
		IncludePharmacistSignature: po.IncludePharmacistSignature,
		CreatedByEmployee: Employee{
			ID:   po.CreatedByEmployeeID,
			Name: po.CreatedByEmployeeName,
		},
		CreatedByEmail: po.CreatedByEmail,
		CreatedAt:      po.CreatedAt,
		Cancellation:   cancellationFromDB(po),
	}
}

func detailFromDB(po models.PurchaseOrder) Detail {
	supplier := po.SupplierSnapshot.Data()

	var signer *Signer
	if s := po.SignerSnapshot.Data(); s != nil {
		signer = &Signer{Name: s.Name, LicenseNumber: s.LicenseNumber}
	}

	items := make([]Item, len(po.Items))
	for i, item := range po.Items {
		items[i] = Item{
			Position:       item.Position,
			DrugVmedisCode: item.DrugVmedisCode,
			DrugName:       item.DrugName,
			Quantity:       item.Quantity,

			ActiveIngredient:   item.ActiveIngredient,
			DosageFormStrength: item.DosageFormStrength,
			SmallestPackage:    item.SmallestPackage,
			QuantityWords:      item.QuantityWords,
			Note:               item.Note,
		}
	}

	return Detail{
		ID:                 po.ID,
		Number:             po.Number,
		TemplateID:         po.TemplateCode,
		TemplateName:       templateName(po.TemplateCode),
		TemplateItemFields: templateItemFields(po.TemplateCode),
		Status:             statusFromDB(po),
		Supplier: Supplier{
			VmedisID: po.SupplierVmedisID,
			Name:     supplier.Name,
			Address:  supplier.Address,
			City:     supplier.City,
			Phone:    supplier.Phone,
		},
		ItemCount:                  len(po.Items),
		IncludePharmacistSignature: po.IncludePharmacistSignature,
		Signer:                     signer,
		CreatedByEmployee: Employee{
			ID:   po.CreatedByEmployeeID,
			Name: po.CreatedByEmployeeName,
		},
		CreatedByEmail: po.CreatedByEmail,
		CreatedAt:      po.CreatedAt,
		Cancellation:   cancellationFromDB(po),
		Items:          items,
	}
}
