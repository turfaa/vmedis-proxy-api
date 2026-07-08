package vmedisv1

import (
	"context"
	"fmt"
	"io"

	"github.com/PuerkitoBio/goquery"
)

// OutOfStockDrugsResponse is the response of the Out-of-Stock Drugs client method.
type OutOfStockDrugsResponse struct {
	Drugs      []DrugStock
	OtherPages []int
}

// GetAllOutOfStockDrugs gets all the out-of-stock drugs from vmedis.
// It fetches every /obathabis-batch/index?page=<page> page concurrently and
// returns an error if any page cannot be fetched or parsed.
func (c *Client) GetAllOutOfStockDrugs(ctx context.Context) ([]DrugStock, error) {
	return getAllPages(ctx, "out-of-stock drugs", c.concurrency, func(ctx context.Context, page int) ([]DrugStock, []int, error) {
		res, err := c.GetOutOfStockDrugs(ctx, page)
		if err != nil {
			return nil, nil, err
		}

		return res.Drugs, res.OtherPages, nil
	})
}

// GetOutOfStockDrugs gets the out-of-stock drugs from vmedis.
// It calls the /obathabis-batch/index?page=<page> page and try to parse the out-of-stock drugs from it.
func (c *Client) GetOutOfStockDrugs(ctx context.Context, page int) (OutOfStockDrugsResponse, error) {
	res, err := c.get(ctx, fmt.Sprintf("/obathabis-batch/index?page=%d", page))
	if err != nil {
		return OutOfStockDrugsResponse{}, fmt.Errorf("get out of stock drugs at page %d: %w", page, err)
	}
	defer res.Body.Close()

	drugs, err := ParseOutOfStockDrugs(res.Body)
	if err != nil {
		return OutOfStockDrugsResponse{}, fmt.Errorf("parse out of stock drugs at page %d: %w", page, err)
	}

	return drugs, nil
}

// ParseOutOfStockDrugs parses the out-of-stock drugs from the given reader.
func ParseOutOfStockDrugs(r io.Reader) (OutOfStockDrugsResponse, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return OutOfStockDrugsResponse{}, fmt.Errorf("parse HTML: %w", err)
	}

	var drugs []DrugStock
	doc.Find("tr[data-key]").EachWithBreak(func(i int, s *goquery.Selection) bool {
		drug, parseErr := parseOutOfStockDrug(s)
		if parseErr != nil {
			err = fmt.Errorf("parse out-of-stock drug #%d: %w", i, parseErr)
			return false
		}

		drugs = append(drugs, drug)
		return true
	})
	if err != nil {
		return OutOfStockDrugsResponse{}, err
	}

	return OutOfStockDrugsResponse{Drugs: drugs, OtherPages: parsePagination(doc)}, nil
}

func parseOutOfStockDrug(doc *goquery.Selection) (DrugStock, error) {
	var ds DrugStock
	if err := UnmarshalDataColumn("oos-column", doc, &ds); err != nil {
		return DrugStock{}, fmt.Errorf("parse drug: %w", err)
	}

	if ds.Drug.MinimumStock.Unit == "" {
		ds.Drug.MinimumStock.Unit = ds.Stock.Unit
	}

	return ds, nil
}
