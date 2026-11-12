package vmedisv2

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// OutOfStockDrug is a drug whose stock is at or below its minimum stock.
type OutOfStockDrug struct {
	VmedisCode   string
	Name         string
	Manufacturer string
	Supplier     string
	MinimumStock Stock
	Stock        Stock
}

// Stock is a quantity of a drug in a unit.
type Stock struct {
	Unit     string
	Quantity float64
}

// outOfStockDrugsPageSize is the number of out-of-stock drugs fetched per page.
const outOfStockDrugsPageSize = 1000

// GetAllOutOfStockDrugs gets all the out-of-stock drugs, like the "Obat Habis"
// v2 report page (/obathabis-batch-v2) lists them: active drugs with the
// out-of-stock notification enabled, per warehouse.
func (c *Client) GetAllOutOfStockDrugs(ctx context.Context) ([]OutOfStockDrug, error) {
	var drugs []OutOfStockDrug
	for page, offset := 1, 0; ; page, offset = page+1, offset+outOfStockDrugsPageSize {
		log.Printf("Getting out-of-stock drugs page %d (%d per page)", page, outOfStockDrugsPageSize)

		var res struct {
			Data []outOfStockDrugRow `json:"data"`
		}
		err := c.callReport(ctx, "lap-obat-habis/list", map[string]any{
			"gudid":          "pergudang",
			"obathabisnotif": "1",
			"obatstatus":     "1",
			"limit":          outOfStockDrugsPageSize,
			"offset":         offset,
		}, &res)
		if err != nil {
			return nil, fmt.Errorf("get out-of-stock drugs page %d: %w", page, err)
		}

		log.Printf("Got %d out-of-stock drugs from page %d", len(res.Data), page)

		for _, row := range res.Data {
			drugs = append(drugs, row.toOutOfStockDrug())
		}

		if len(res.Data) < outOfStockDrugsPageSize {
			break
		}
	}

	log.Printf("Got %d out-of-stock drugs", len(drugs))

	return drugs, nil
}

// outOfStockDrugRow is a row of lap-obat-habis/list.
type outOfStockDrugRow struct {
	VmedisCode   string `json:"obatkode"`
	Name         string `json:"obatnama"`
	Manufacturer string `json:"pabnama"`
	Supplier     string `json:"supnama"`
	MinimumStock number `json:"obatminstok"`
	Stock        number `json:"stok"`
	// Unit is the unit of both the stock and the minimum stock: the drug's
	// smallest unit.
	Unit string `json:"sonama"`
}

func (r outOfStockDrugRow) toOutOfStockDrug() OutOfStockDrug {
	unit := strings.TrimSpace(r.Unit)

	return OutOfStockDrug{
		VmedisCode:   r.VmedisCode,
		Name:         strings.TrimSpace(r.Name),
		Manufacturer: strings.TrimSpace(r.Manufacturer),
		Supplier:     strings.TrimSpace(r.Supplier),
		MinimumStock: Stock{Unit: unit, Quantity: float64(r.MinimumStock)},
		Stock:        Stock{Unit: unit, Quantity: float64(r.Stock)},
	}
}
