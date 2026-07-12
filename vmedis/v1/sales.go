package vmedisv1

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// SalesResponse is the response of Sales client method.
type SalesResponse struct {
	Sales      []Sale
	OtherPages []int
}

// GetAllTodaySales gets all sales from today from vmedis.
// It fetches every /apt-lap-penjualanobat-batch/index?page=<page> page
// concurrently and returns an error if any page cannot be fetched or parsed.
func (c *Client) GetAllTodaySales(ctx context.Context) ([]Sale, error) {
	return getAllPages(ctx, "today sales", c.concurrency, func(ctx context.Context, page int) ([]Sale, []int, error) {
		res, err := c.GetTodaySales(ctx, page)
		if err != nil {
			return nil, nil, err
		}

		return res.Sales, res.OtherPages, nil
	})
}

// GetAllSalesBetweenDates gets all sales between the given dates from vmedis.
// It fetches every /apt-lap-penjualanobat-batch/index page concurrently
// and returns an error if any page cannot be fetched or parsed.
func (c *Client) GetAllSalesBetweenDates(ctx context.Context, startDate time.Time, endDate time.Time) ([]Sale, error) {
	return getAllPages(ctx, "sales", c.concurrency, func(ctx context.Context, page int) ([]Sale, []int, error) {
		res, err := c.GetSales(ctx, SearchByTimeParameters[ParameterTypeSales]{
			StartTime: startDate,
			EndTime:   endDate,
			Page:      page,
		})
		if err != nil {
			return nil, nil, err
		}

		return res.Sales, res.OtherPages, nil
	})
}

// GetTodaySales gets one page of the sales from today from vmedis.
// It calls the /apt-lap-penjualanobat-batch/index?page=<page> page and try to parse the sales from it.
func (c *Client) GetTodaySales(ctx context.Context, page int) (SalesResponse, error) {
	return c.GetSales(ctx, SearchByTimeParameters[ParameterTypeSales]{Page: page})
}

// GetSales gets one page of sales matching the given search parameters from vmedis.
// It calls the /apt-lap-penjualanobat-batch/index page and tries to parse the sales from it.
func (c *Client) GetSales(ctx context.Context, params SearchByTimeParameters[ParameterTypeSales]) (SalesResponse, error) {
	res, err := c.get(ctx, fmt.Sprintf("/apt-lap-penjualanobat-batch/index?%s", params.ToQuery(dateFormat)))
	if err != nil {
		return SalesResponse{}, fmt.Errorf("get sales with params %+v: %w", params, err)
	}
	defer res.Body.Close()

	sales, err := ParseSales(res.Body)
	if err != nil {
		return SalesResponse{}, fmt.Errorf("parse sales with params %+v: %w", params, err)
	}

	return sales, nil
}

// ParseSales parses the sales from the given reader.
// It usually comes from the /apt-lap-penjualanobat-batch/index?page=<page> page.
func ParseSales(r io.Reader) (SalesResponse, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return SalesResponse{}, fmt.Errorf("new document from reader: %w", err)
	}

	var sales []Sale
	doc.Find("tr[data-key]").EachWithBreak(func(i int, s *goquery.Selection) bool {
		sale, parseErr := parseSale(s)
		if parseErr != nil {
			err = fmt.Errorf("parse sale #%d: %w", i, parseErr)
			return false
		}

		sales = append(sales, sale)
		return true
	})
	if err != nil {
		return SalesResponse{}, err
	}

	return SalesResponse{Sales: sales, OtherPages: parsePagination(doc)}, nil
}

func parseSale(selection *goquery.Selection) (Sale, error) {
	var sale Sale
	if err := UnmarshalDataColumn("sales-column", selection, &sale); err != nil {
		return Sale{}, fmt.Errorf("unmarshal sale: %w", err)
	}

	var err error
	selection.Find("table tr:nth-child(n+2)").EachWithBreak(func(i int, s *goquery.Selection) bool {
		su, parseErr := parseSaleUnit(s)
		if parseErr != nil {
			err = fmt.Errorf("parse sale unit #%d: %w", i, parseErr)
			return false
		}

		sale.SaleUnits = append(sale.SaleUnits, su)
		return true
	})
	if err != nil {
		return Sale{}, err
	}

	// Get the value from <button type="button" class="btn btn-warning btn-xs actionPrint" value="110844" title="Cetak Faktur">.
	idStr, ok := selection.Find("button.actionPrint").Attr("value")
	if ok {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return Sale{}, fmt.Errorf("parse sale vmedis id: %w", err)
		}

		sale.ID = id
	} else {
		html, _ := selection.Html()
		return Sale{}, fmt.Errorf("sale vmedis id not found in: %s", html)
	}

	return sale, nil
}

func parseSaleUnit(selection *goquery.Selection) (SaleUnit, error) {
	var su SaleUnit
	if err := UnmarshalDataColumnByIndex("sales-index", selection, &su); err != nil {
		return SaleUnit{}, fmt.Errorf("unmarshal sale unit: %w", err)
	}

	return su, nil
}
