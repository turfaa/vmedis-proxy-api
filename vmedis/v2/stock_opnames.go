package vmedisv2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// StockOpname is one drug batch counted in a stock opname (inventory count).
type StockOpname struct {
	ID int

	// Code is the stock opname code, e.g. "SOP202610020001". It is not unique:
	// vmedis gives several counted batches the same code.
	Code string

	// Time is when the batch was counted.
	Time time.Time

	DrugCode  string
	DrugName  string
	Unit      string
	Warehouse string

	BatchCode string
	// ExpiryDate is zero when the batch has no expiry date.
	ExpiryDate time.Time

	// InitialQuantity is the stock vmedis had before the count.
	InitialQuantity float64
	// RealQuantity is the counted stock.
	RealQuantity float64
	// QuantityDifference is RealQuantity minus InitialQuantity.
	QuantityDifference float64

	// HPP is the unit cost (harga pokok penjualan) of the drug at the count.
	HPP float64
	// SalePrice is the unit sale price (harga jual 1) of the drug at the count.
	SalePrice float64

	// HPPDifference is QuantityDifference valued at HPP, the "Total HPP" the
	// report page shows.
	HPPDifference float64
	// SalePriceDifference is QuantityDifference valued at SalePrice, the
	// "Total Harga Jual" the report page shows.
	SalePriceDifference float64

	Notes    string
	Operator string
}

// stockOpnamesPageSize is the number of stock opnames fetched per page.
const stockOpnamesPageSize = 500

// stockOpnamesQueryDateLayout is the layout of the tglAwal and tglAkhir
// filters, e.g. "02 Oct 2026".
const stockOpnamesQueryDateLayout = "02 Jan 2006"

// GetAllStockOpnamesBetweenDates gets all the stock opnames counted between
// the given dates, inclusive, ordered by ID.
// It calls the endpoint the "Lap Stok Opname" v2 report page
// (/apt-lap-stok-opname-batch-v2) gets its rows from.
func (c *Client) GetAllStockOpnamesBetweenDates(ctx context.Context, startDate, endDate time.Time) ([]StockOpname, error) {
	dates := fmt.Sprintf("%s - %s", startDate.Format(stockOpnamesQueryDateLayout), endDate.Format(stockOpnamesQueryDateLayout))

	var stockOpnames []StockOpname
	for page, offset := 1, 0; ; page, offset = page+1, offset+stockOpnamesPageSize {
		log.Printf("Getting stock opnames %s page %d (%d per page)", dates, page, stockOpnamesPageSize)

		rows, err := c.getStockOpnamesPage(ctx, startDate, endDate, offset)
		if err != nil {
			return nil, fmt.Errorf("get stock opnames %s page %d: %w", dates, page, err)
		}

		log.Printf("Got %d stock opnames from %s page %d", len(rows), dates, page)

		for _, row := range rows {
			stockOpname, err := row.toStockOpname()
			if err != nil {
				return nil, fmt.Errorf("convert stock opname %d (%s): %w", int(row.ID), row.Code, err)
			}
			stockOpnames = append(stockOpnames, stockOpname)
		}

		if len(rows) < stockOpnamesPageSize {
			break
		}
	}

	log.Printf("Got %d stock opnames %s", len(stockOpnames), dates)

	return stockOpnames, nil
}

func (c *Client) getStockOpnamesPage(ctx context.Context, startDate, endDate time.Time, offset int) ([]stockOpnameRow, error) {
	var res struct {
		Data []stockOpnameRow `json:"data"`
	}
	err := c.callReport(ctx, "apt-lap-stok-opname-batch/index", map[string]any{
		// cari 4 filters by tglAwal and tglAkhir.
		"cari":     4,
		"tglAwal":  startDate.Format(stockOpnamesQueryDateLayout),
		"tglAkhir": endDate.Format(stockOpnamesQueryDateLayout),
		// Without a stable order, the pages could overlap or skip stock opnames.
		"sorting": "sopid ASC",
		"limit":   stockOpnamesPageSize,
		"offset":  offset,
	}, &res)
	return res.Data, err
}

// stockOpnameRow is a row of apt-lap-stok-opname-batch/index.
type stockOpnameRow struct {
	ID   number `json:"sopid"`
	Code string `json:"sopkode"`
	// Time is like "2026-10-02T09:13:34.000Z". Despite the "Z", it is a local
	// time.
	Time      string  `json:"soptanggal"`
	DrugCode  string  `json:"obatkode"`
	DrugName  string  `json:"obatnama"`
	Unit      string  `json:"sonama"`
	Warehouse *string `json:"gudnama"`
	BatchCode *string `json:"sopnobatch"`
	// ExpiryDate is like "2026-10-01T00:00:00.000Z". Despite the "Z", it is a
	// local date.
	ExpiryDate *string `json:"soptglexpired"`

	InitialQuantity    number `json:"stok"`
	RealQuantity       number `json:"sopjumlah"`
	QuantityDifference number `json:"sopselisih"`

	HPP       number `json:"sophrgpp"`
	SalePrice number `json:"sophrgjual1"`

	Notes    *string `json:"sopketerangan"`
	Operator *string `json:"username"`
}

// stockOpnameTimeLayout is the layout of soptanggal and soptglexpired.
const stockOpnameTimeLayout = "2006-01-02T15:04:05.000Z"

func (r stockOpnameRow) toStockOpname() (StockOpname, error) {
	t, err := time.ParseInLocation(stockOpnameTimeLayout, r.Time, time.Local)
	if err != nil {
		return StockOpname{}, fmt.Errorf("parse time %q: %w", r.Time, err)
	}

	var expiryDate time.Time
	if r.ExpiryDate != nil && *r.ExpiryDate != "" {
		expiryDate, err = time.ParseInLocation(stockOpnameTimeLayout, *r.ExpiryDate, time.Local)
		if err != nil {
			return StockOpname{}, fmt.Errorf("parse expiry date %q: %w", *r.ExpiryDate, err)
		}
	}

	quantityDifference := float64(r.QuantityDifference)

	return StockOpname{
		ID:                  int(r.ID),
		Code:                strings.TrimSpace(r.Code),
		Time:                t,
		DrugCode:            strings.TrimSpace(r.DrugCode),
		DrugName:            strings.TrimSpace(r.DrugName),
		Unit:                strings.TrimSpace(r.Unit),
		Warehouse:           strings.TrimSpace(deref(r.Warehouse)),
		BatchCode:           strings.TrimSpace(deref(r.BatchCode)),
		ExpiryDate:          expiryDate,
		InitialQuantity:     float64(r.InitialQuantity),
		RealQuantity:        float64(r.RealQuantity),
		QuantityDifference:  quantityDifference,
		HPP:                 float64(r.HPP),
		SalePrice:           float64(r.SalePrice),
		HPPDifference:       quantityDifference * float64(r.HPP),
		SalePriceDifference: quantityDifference * float64(r.SalePrice),
		Notes:               strings.TrimSpace(deref(r.Notes)),
		Operator:            strings.TrimSpace(deref(r.Operator)),
	}, nil
}
