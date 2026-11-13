package vmedisv2

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
)

// Drug is a drug in vmedis.
type Drug struct {
	VmedisID     int64
	VmedisCode   string
	KFACode      string
	Name         string
	Manufacturer string

	// MinimumStock is in the drug's smallest unit.
	// Only set by GetDrug.
	MinimumStock Stock

	// Units are ordered from the smallest unit.
	// Only set by GetDrug.
	Units []Unit

	// Stocks are the drug's stock per unit, ordered from the biggest unit.
	// Only set by GetDrug.
	Stocks []Stock
}

// Unit is a unit of a drug.
type Unit struct {
	Unit string

	// ParentUnit is the next smaller unit, empty for the smallest unit.
	ParentUnit string

	// ConversionToParentUnit is how many of the parent unit make one of this
	// unit, 1 for the smallest unit.
	ConversionToParentUnit float64

	// UnitOrder is the 0-based position of the unit, from the smallest.
	UnitOrder int

	// PriceOne is the price for common customers, PriceTwo for medical
	// facilities, and PriceThree for prescriptions.
	PriceOne   float64
	PriceTwo   float64
	PriceThree float64
}

// drugsPageSize is the number of drugs fetched per page. The endpoint returns
// at most 100 per page, whatever limit it is asked for.
const drugsPageSize = 100

// GetAllDrugs gets all the active drugs, without their units and stocks, from
// the endpoint the "Data Obat" v2 page (/apt-obat-batch-v2) lists them from.
func (c *Client) GetAllDrugs(ctx context.Context) ([]Drug, error) {
	var drugs []Drug
	for page := 1; ; page++ {
		log.Printf("Getting drugs page %d (%d per page)", page, drugsPageSize)

		var res struct {
			Status  int       `json:"status"`
			Message string    `json:"message"`
			Data    []drugRow `json:"data"`
			Total   number    `json:"total"`
		}
		err := c.callReport(ctx, "apt-obat-batch/index", map[string]any{
			"obatstatus": "1",
			// Every warehouse.
			"gudid": "",
			// A stable order, so that the pages do not overlap or skip drugs.
			"sorting": "obatid ASC",
			"page":    page,
			"limit":   drugsPageSize,
		}, &res)
		if err != nil {
			return nil, fmt.Errorf("get drugs page %d: %w", page, err)
		}
		if res.Status != 1 {
			return nil, fmt.Errorf("get drugs page %d: status %d: %s", page, res.Status, res.Message)
		}

		log.Printf("Got %d drugs from page %d, %d of %d", len(res.Data), page, len(drugs)+len(res.Data), int(res.Total))

		for _, row := range res.Data {
			drugs = append(drugs, row.toDrug())
		}

		// Stop at the total, or at an empty page in case the total is off.
		if len(res.Data) == 0 || len(drugs) >= int(res.Total) {
			break
		}
	}

	if len(drugs) == 0 {
		return nil, fmt.Errorf("got no drugs")
	}

	log.Printf("Got %d drugs", len(drugs))

	return drugs, nil
}

// GetDrug gets the drug with the given vmedis ID, with its units and stocks,
// from the endpoint the "Data Obat" v2 page shows a drug's details from.
func (c *Client) GetDrug(ctx context.Context, id int64) (Drug, error) {
	var res struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Drug  drugRow   `json:"obat"`
			Units []unitRow `json:"satuan"`
			Cart  []cartRow `json:"cart"`
		} `json:"data"`
	}
	err := c.callReport(ctx, "apt-obat-batch/view", map[string]any{
		"obatid": id,
		// Every warehouse.
		"gudid": "",
	}, &res)
	if err != nil {
		return Drug{}, fmt.Errorf("get drug %d: %w", id, err)
	}
	if res.Status != 1 {
		return Drug{}, fmt.Errorf("get drug %d: status %d: %s", id, res.Status, res.Message)
	}

	drug := res.Data.Drug.toDrug()
	drug.Units = toUnits(res.Data.Units)
	drug.Stocks = toStocks(res.Data.Cart, drug.Units)

	if len(drug.Units) > 0 {
		drug.MinimumStock.Unit = drug.Units[0].Unit
	}

	return drug, nil
}

// drugRow is a row of apt-obat-batch/index, and the drug of apt-obat-batch/view.
type drugRow struct {
	ID           number  `json:"obatid"`
	Code         string  `json:"obatkode"`
	KFACode      *string `json:"kodekfa"`
	Name         string  `json:"obatnama"`
	Manufacturer *string `json:"pabnama"`
	MinimumStock number  `json:"obatminstok"`
}

func (r drugRow) toDrug() Drug {
	return Drug{
		VmedisID:     int64(r.ID),
		VmedisCode:   strings.TrimSpace(r.Code),
		KFACode:      strings.TrimSpace(deref(r.KFACode)),
		Name:         strings.TrimSpace(r.Name),
		Manufacturer: strings.TrimSpace(deref(r.Manufacturer)),
		MinimumStock: Stock{Quantity: float64(r.MinimumStock)},
	}
}

// unitRow is a unit in apt-obat-batch/view.
type unitRow struct {
	Name string `json:"sonama"`
	// Order is the 1-based position of the unit, from the smallest.
	Order number `json:"sodurut"`
	// Conversion is how many of the previous unit make one of this unit.
	Conversion number `json:"sodkonversi"`
	PriceOne   number `json:"hrgjual1"`
	PriceTwo   number `json:"hrgjual2"`
	PriceThree number `json:"hrgjual3"`
}

func toUnits(rows []unitRow) []Unit {
	rows = slices.Clone(rows)
	slices.SortFunc(rows, func(a, b unitRow) int { return cmp.Compare(a.Order, b.Order) })

	units := make([]Unit, 0, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}

		unit := Unit{
			Unit:                   name,
			ConversionToParentUnit: float64(row.Conversion),
			UnitOrder:              len(units),
			PriceOne:               float64(row.PriceOne),
			PriceTwo:               float64(row.PriceTwo),
			PriceThree:             float64(row.PriceThree),
		}
		if len(units) > 0 {
			unit.ParentUnit = units[len(units)-1].Unit
		}

		units = append(units, unit)
	}

	return units
}

// cartRow is a batch of the drug's stock in apt-obat-batch/view.
type cartRow struct {
	Unit  string `json:"sonama"`
	Stock number `json:"stok"`
}

// toStocks sums the batches' stock per unit, leaving out the empty ones, and
// orders them from the biggest unit.
func toStocks(rows []cartRow, units []Unit) []Stock {
	var stocks []Stock
	for _, row := range rows {
		unit := strings.TrimSpace(row.Unit)
		if row.Stock == 0 {
			continue
		}

		i := slices.IndexFunc(stocks, func(s Stock) bool { return s.Unit == unit })
		if i < 0 {
			stocks = append(stocks, Stock{Unit: unit})
			i = len(stocks) - 1
		}
		stocks[i].Quantity += float64(row.Stock)
	}

	order := func(unit string) int {
		i := slices.IndexFunc(units, func(u Unit) bool { return u.Unit == unit })
		return i
	}
	slices.SortStableFunc(stocks, func(a, b Stock) int { return cmp.Compare(order(b.Unit), order(a.Unit)) })

	return stocks
}
