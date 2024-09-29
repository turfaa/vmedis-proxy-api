package drug

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/turfaa/vmedis-proxy-api/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis"
)

// DrugsResponse is the response for the drugs endpoints.
type DrugsResponse struct {
	Drugs []Drug `json:"drugs"`
}

type WithStock struct {
	Drug  Drug  `json:"drug"`
	Stock Stock `json:"stock"`
}

func FromVmedisDrugStock(ds vmedis.DrugStock) WithStock {
	return WithStock{
		Drug:  FromVmedisDrug(ds.Drug),
		Stock: FromVmedisStock(ds.Stock),
	}
}

// Drug is a drug in the inventory.
type Drug struct {
	VmedisCode   string  `json:"vmedisCode,omitempty"`
	Name         string  `json:"name,omitempty"`
	Manufacturer string  `json:"manufacturer,omitempty"`
	Supplier     string  `json:"supplier,omitempty"`
	MinimumStock Stock   `json:"minimumStock"`
	Units        []Unit  `json:"units"`
	Stocks       []Stock `json:"stocks"`
}

// FromVmedisDrug creates Drug from its vmedis schema.
func FromVmedisDrug(cd vmedis.Drug) Drug {
	units := make([]Unit, 0, len(cd.Units))
	for _, cu := range cd.Units {
		units = append(units, FromVmedisUnit(cu))
	}

	stocks := make([]Stock, 0, len(cd.Stocks))
	for _, cs := range cd.Stocks {
		stocks = append(stocks, FromVmedisStock(cs))
	}

	return Drug{
		VmedisCode:   cd.VmedisCode,
		Name:         cd.Name,
		Manufacturer: cd.Manufacturer,
		Supplier:     cd.Supplier,
		MinimumStock: FromVmedisStock(cd.MinimumStock),
		Units:        units,
		Stocks:       stocks,
	}
}

// FromDBDrug creates Drug from models.Drug.
func FromDBDrug(drug models.Drug) Drug {
	units := make([]Unit, 0, len(drug.Units))
	for _, mu := range drug.Units {
		units = append(units, FromDBDrugUnit(mu))
	}

	stocks := make([]Stock, 0, len(drug.Stocks))
	for _, ms := range drug.Stocks {
		stocks = append(stocks, FromDBDrugStock(ms))
	}

	return Drug{
		VmedisCode:   drug.VmedisCode,
		Name:         drug.Name,
		Manufacturer: drug.Manufacturer,
		MinimumStock: FromDBStock(drug.MinimumStock),
		Units:        units,
		Stocks:       stocks,
	}
}

// Stock represents one instance of stock.
type Stock struct {
	Unit     string  `json:"unit"`
	Quantity float64 `json:"quantity"`
}

func (s Stock) String() string {
	b, err := s.MarshalText()
	if err != nil {
		return fmt.Sprintf("%0.0f %s", s.Quantity, s.Unit)
	}

	return string(b)
}

// MarshalText implements encoding.TextMarshaler.
func (s Stock) MarshalText() ([]byte, error) {
	q, err := json.Marshal(s.Quantity)
	if err != nil {
		return nil, fmt.Errorf("marshal quantity: %w", err)
	}

	var b bytes.Buffer
	b.Write(q)
	if s.Unit != "" {
		b.WriteByte(' ')
		b.WriteString(s.Unit)
	}

	return bytes.TrimSpace(b.Bytes()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (s *Stock) UnmarshalText(text []byte) error {
	split := bytes.SplitN(bytes.TrimSpace(text), []byte(" "), 2)

	var q float64
	if err := json.Unmarshal(split[0], &q); err != nil {
		return fmt.Errorf("unmarshal quantity: %w", err)
	}

	s.Quantity = q
	if len(split) > 1 {
		s.Unit = string(split[1])
	}

	return nil
}

// FromVmedisStock creates Stock from its vmedis schema.
func FromVmedisStock(cs vmedis.Stock) Stock {
	return Stock{
		Unit:     cs.Unit,
		Quantity: cs.Quantity,
	}
}

// FromDBDrugStock creates Stock from models.DrugStock.
func FromDBDrugStock(stock models.DrugStock) Stock {
	return FromDBStock(stock.Stock)
}

// FromDBStock creates Stock from models.Stock.
func FromDBStock(stock models.Stock) Stock {
	return Stock{
		Unit:     stock.Unit,
		Quantity: stock.Quantity,
	}
}

// Unit is a unit of a drug.
type Unit struct {
	Unit                   string  `json:"unit,omitempty"`
	ParentUnit             string  `json:"parentUnit,omitempty"`
	ConversionToParentUnit float64 `json:"conversionToParentUnit,omitempty"`

	// UnitOrder is the order of the unit of the drug.
	// The smallest unit has the lowest order.
	UnitOrder int `json:"unitOrder"`

	// PriceOne, PriceTwo, and PriceThree are the prices of the drug for different segments.
	// PriceOne is the price for common customers.
	// PriceTwo is the price for medical facilities.
	// PriceThree is the price for prescription.
	PriceOne   float64 `json:"priceOne,omitempty"`
	PriceTwo   float64 `json:"priceTwo,omitempty"`
	PriceThree float64 `json:"priceThree,omitempty"`
}

// FromVmedisUnit creates Unit from its vmedis schema.
func FromVmedisUnit(cu vmedis.Unit) Unit {
	return Unit{
		Unit:                   cu.Unit,
		ParentUnit:             cu.ParentUnit,
		ConversionToParentUnit: cu.ConversionToParentUnit,
		UnitOrder:              cu.UnitOrder,
		PriceOne:               cu.PriceOne,
		PriceTwo:               cu.PriceTwo,
		PriceThree:             cu.PriceThree,
	}
}

// FromDBDrugUnit creates Unit from models.DrugUnit.
func FromDBDrugUnit(unit models.DrugUnit) Unit {
	return Unit{
		Unit:                   unit.Unit,
		ParentUnit:             unit.ParentUnit,
		ConversionToParentUnit: unit.ConversionToParentUnit,
		UnitOrder:              unit.UnitOrder,
		PriceOne:               unit.PriceOne,
		PriceTwo:               unit.PriceTwo,
		PriceThree:             unit.PriceThree,
	}
}

// SaleStatistics represents statistics of sales of a drug.
type SaleStatistics struct {
	DrugCode      string
	NumberOfSales int
	TotalAmount   float64
}
