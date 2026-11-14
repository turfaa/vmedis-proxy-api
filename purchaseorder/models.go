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
}

// TemplateInfo describes a purchase order template.
type TemplateInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
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
	IdempotencyKey             string       `json:"idempotencyKey"`
	TemplateID                 string       `json:"templateId"`
	SupplierVmedisID           int          `json:"supplierVmedisId"`
	IncludePharmacistSignature bool         `json:"includePharmacistSignature"`
	CreatedByEmployee          Employee     `json:"createdByEmployee"`
	Items                      []CreateItem `json:"items"`
}

// CreateItem is an item of the purchase order to create.
type CreateItem struct {
	// DrugVmedisCode is null for drugs added by hand that are not in vmedis.
	DrugVmedisCode *string `json:"drugVmedisCode"`
	DrugName       string  `json:"drugName"`
	Quantity       string  `json:"quantity"`
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

// Summary is a purchase order in the list.
type Summary struct {
	ID                         uint            `json:"id"`
	Number                     string          `json:"number"`
	TemplateID                 string          `json:"templateId"`
	TemplateName               string          `json:"templateName"`
	Supplier                   SummarySupplier `json:"supplier"`
	ItemCount                  int             `json:"itemCount"`
	IncludePharmacistSignature bool            `json:"includePharmacistSignature"`
	CreatedByEmployee          Employee        `json:"createdByEmployee"`
	CreatedByEmail             string          `json:"createdByEmail"`
	CreatedAt                  time.Time       `json:"createdAt"`
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
}

// Detail is a purchase order with its items.
type Detail struct {
	ID                         uint      `json:"id"`
	Number                     string    `json:"number"`
	TemplateID                 string    `json:"templateId"`
	TemplateName               string    `json:"templateName"`
	Supplier                   Supplier  `json:"supplier"`
	ItemCount                  int       `json:"itemCount"`
	IncludePharmacistSignature bool      `json:"includePharmacistSignature"`
	Signer                     *Signer   `json:"signer"`
	CreatedByEmployee          Employee  `json:"createdByEmployee"`
	CreatedByEmail             string    `json:"createdByEmail"`
	CreatedAt                  time.Time `json:"createdAt"`
	Items                      []Item    `json:"items"`
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

func summaryFromDB(po models.PurchaseOrder, itemCount int) Summary {
	supplier := po.SupplierSnapshot.Data()

	return Summary{
		ID:           po.ID,
		Number:       po.Number,
		TemplateID:   po.TemplateCode,
		TemplateName: templateName(po.TemplateCode),
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
		}
	}

	return Detail{
		ID:           po.ID,
		Number:       po.Number,
		TemplateID:   po.TemplateCode,
		TemplateName: templateName(po.TemplateCode),
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
		Items:          items,
	}
}
