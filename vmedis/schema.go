package vmedis

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// DrugStock is the stock of a drug.
type DrugStock struct {
	Drug  Drug  `oos-column:"<self>"`
	Stock Stock `oos-column:"8"`
}

// Drug is a drug in the inventory.
type Drug struct {
	VmedisID     int64
	VmedisCode   string `oos-column:"4" drugs-index:"3" form-name:"Obat[obatkode]"`
	Name         string `oos-column:"5" drugs-index:"4" form-name:"Obat[obatnama]"`
	Manufacturer string `oos-column:"12" drugs-index:"5" form-name:"Obat[pabid]"`
	Supplier     string `oos-column:"13"`
	MinimumStock Stock  `oos-column:"6" form-name:"Obat[obatminstok]"`
	Units        []Unit `form-name:"Obat[soid%d]"`
	Stocks       []Stock
}

// Stock represents one instance of stock.
type Stock struct {
	Unit     string
	Quantity float64
}

// UnmarshalDataColumn implements DataColumnUnmarshaler.
func (s *Stock) UnmarshalDataColumn(selection *goquery.Selection) error {
	stockString := selection.Text()
	return s.UnmarshalText([]byte(stockString))
}

// UnmarshalForm implements FormUnmarshaler.
func (s *Stock) UnmarshalForm(selection *goquery.Selection) error {
	stockString := selection.AttrOr("value", "")
	return s.UnmarshalText([]byte(stockString))
}

// UnmarshalText implements TextUnmarshaler.
func (s *Stock) UnmarshalText(text []byte) error {
	stockString := string(text)
	stockString = strings.TrimSpace(stockString)

	split := strings.Split(stockString, " ")
	if len(split) > 0 {
		q, err := parseFloat(split[0])
		if err != nil {
			return fmt.Errorf("parse stock quantity from string [%s]: %w", stockString, err)
		}

		s.Quantity = q
	}

	if len(split) > 1 {
		s.Unit = split[1]
	}

	return nil
}

// Unit represents a unit of a drug.
type Unit struct {
	Unit                   string
	ParentUnit             string
	ConversionToParentUnit float64

	// UnitOrder is the order of the unit of the drug.
	// The smallest unit has the lowest order.
	UnitOrder int

	// PriceOne, PriceTwo, and PriceThree are the prices of the drug for different segments.
	// PriceOne is the price for common customers.
	// PriceTwo is the price for medical facilities.
	// PriceThree is the price for prescription.
	PriceOne   float64
	PriceTwo   float64
	PriceThree float64

	formName string
}

// UnmarshalForm implements FormUnmarshaler.
// It only parses the unit name.
// To complete the unit data, use the EnrichUnitsFromDoc function.
func (u *Unit) UnmarshalForm(selection *goquery.Selection) error {
	u.formName = selection.AttrOr("name", "")
	u.Unit = selection.AttrOr("value", "")
	return nil
}

// EnrichUnitsFromDoc completes the units data by using the data in the goquery document.
// It also filters out the units that doesn't have a form name or a name.
// The document usually comes from the /obat-batch/view?id=<id> page.
func EnrichUnitsFromDoc(units []Unit, doc *goquery.Document) ([]Unit, error) {
	result := make([]Unit, 0, len(units))

	for _, u := range units {
		if u.formName == "" || u.Unit == "" {
			continue
		}

		selection := doc.Selection

		unitNameSelection := selection.Find("input[name=\"" + u.formName + "\"]")
		if unitNameSelection.Length() != 0 {
			u.Unit = unitNameSelection.AttrOr("value", "")
		}

		if len(result) > 0 {
			u.ParentUnit = result[len(result)-1].Unit
		}

		conversionSelection := selection.Find("input[name=\"" + strings.ReplaceAll(u.formName, "soid", "sodkonversi") + "\"]")
		if conversionSelection.Length() > 0 {
			conversionStr := conversionSelection.AttrOr("value", "")
			conversion, err := strconv.ParseFloat(conversionStr, 64)
			if err != nil {
				html, _ := conversionSelection.Html()
				return nil, fmt.Errorf("parse '%s' conversion from string [%s] <%s>: %w", u.Unit, conversionStr, html, err)
			}

			u.ConversionToParentUnit = conversion
		}

		conversion, err := u.extractFloatFromInput(selection, "sodkonversi")
		if err != nil {
			return nil, fmt.Errorf("extract '%s' conversion: %w", u.Unit, err)
		}
		u.ConversionToParentUnit = conversion

		priceOne, err := u.extractFloatFromInput(selection, "hrgjual1")
		if err != nil {
			return nil, fmt.Errorf("extract '%s' price one: %w", u.Unit, err)
		}
		u.PriceOne = priceOne

		priceTwo, err := u.extractFloatFromInput(selection, "hrgjual2")
		if err != nil {
			return nil, fmt.Errorf("extract '%s' price two: %w", u.Unit, err)
		}
		u.PriceTwo = priceTwo

		priceThree, err := u.extractFloatFromInput(selection, "hrgjual3")
		if err != nil {
			return nil, fmt.Errorf("extract '%s' price three: %w", u.Unit, err)
		}
		u.PriceThree = priceThree

		u.UnitOrder = len(result)

		result = append(result, u)
	}

	return result, nil
}

func (u *Unit) extractFloatFromInput(selection *goquery.Selection, name string) (float64, error) {
	selection = selection.Find("input[name=\"" + strings.ReplaceAll(u.formName, "soid", name) + "\"]")
	if selection.Length() == 0 {
		return 0, nil
	}

	floatStr := selection.AttrOr("value", "")
	float, err := strconv.ParseFloat(floatStr, 64)
	if err != nil {
		html, _ := selection.Html()
		return 0, fmt.Errorf("parse '%s' float(%s) from string [%s] <%s>: %w", u.Unit, name, floatStr, html, err)
	}

	return float, nil
}

// Sale represents a sale.
type Sale struct {
	ID            int
	Date          Time    `sales-column:"2"`
	InvoiceNumber string  `sales-column:"6"`
	PatientName   string  `sales-column:"11"`
	Doctor        string  `sales-column:"12"`
	Payment       string  `sales-column:"14"`
	Total         float64 `sales-column:"24"`
	SaleUnits     []SaleUnit
}

// SaleUnit represents one unit of a drug in a sale.
type SaleUnit struct {
	IDInSale      int     `sales-index:"1"`
	DrugCode      string  `sales-index:"4"`
	DrugName      string  `sales-index:"5"`
	Batch         string  `sales-index:"6"`
	Amount        float64 `sales-index:"7"`
	Unit          string  `sales-index:"8"`
	UnitPrice     float64 `sales-index:"9"`
	PriceCategory string  `sales-index:"10"`
	Discount      float64 `sales-index:"12"`
	Tuslah        float64 `sales-index:"13"`
	Embalase      float64 `sales-index:"14"`
	Total         float64 `sales-index:"15"`
}

// StockOpname represents a stock opname.
type StockOpname struct {
	ID                  string  `so-index:"2"`
	Date                Date    `so-index:"3"`
	DrugCode            string  `so-index:"4"`
	DrugName            string  `so-index:"6"`
	BatchCode           string  `so-index:"27"`
	Unit                string  `so-index:"7"`
	InitialQuantity     float64 `so-index:"13"`
	RealQuantity        float64 `so-index:"14"`
	QuantityDifference  float64 `so-index:"15"`
	HPPDifference       float64 `so-index:"24"`
	SalePriceDifference float64 `so-index:"25"`
	Notes               string  `so-index:"26"`
}

type Procurement struct {
	Date                   Date       `procurement-column:"1"`
	InvoiceNumber          string     `procurement-column:"3"`
	Supplier               string     `procurement-column:"4"`
	Warehouse              string     `procurement-column:"9"`
	PaymentType            string     `procurement-column:"10"`
	Operator               string     `procurement-column:"11"`
	CashDiscountPercentage Percentage `procurement-column:"12"`
	DiscountPercentage     Percentage `procurement-column:"13"`
	DiscountAmount         float64    `procurement-column:"14"`
	TaxPercentage          Percentage `procurement-column:"15"`
	TaxAmount              float64    `procurement-column:"16"`
	MiscellaneousCost      float64    `procurement-column:"17"`
	Total                  float64    `procurement-column:"18"`
	ProcurementUnits       []ProcurementUnit
}

type ProcurementUnit struct {
	IDInProcurement         int        `procurement-index:"1"`
	DrugCode                string     `procurement-index:"2"`
	DrugName                string     `procurement-index:"3"`
	Amount                  float64    `procurement-index:"4"`
	Unit                    string     `procurement-index:"5"`
	UnitBasePrice           float64    `procurement-index:"6"`
	DiscountPercentage      Percentage `procurement-index:"7"`
	DiscountTwoPercentage   Percentage `procurement-index:"8"`
	DiscountThreePercentage Percentage `procurement-index:"9"`
	TotalUnitPrice          float64    `procurement-index:"10"`
	UnitTaxedPrice          float64    `procurement-index:"11"`
	ExpiryDate              Date       `procurement-index:"12"`
	BatchNumber             string     `procurement-index:"13"`
	Total                   float64    `procurement-index:"14"`
}
