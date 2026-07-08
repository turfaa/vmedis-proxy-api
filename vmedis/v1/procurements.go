package vmedisv1

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ProcurementsResponse struct {
	Procurements []Procurement
	OtherPages   []int
}

// GetAllProcurementsBetweenDates gets all the procurements between the given dates from vmedis.
// It fetches every /laporan-transaksi-pembelian-obat-batch/index page concurrently
// and returns an error if any page cannot be fetched or parsed.
func (c *Client) GetAllProcurementsBetweenDates(
	ctx context.Context,
	startDate time.Time,
	endDate time.Time,
) ([]Procurement, error) {
	return getAllPages(ctx, "procurements", c.concurrency, func(ctx context.Context, page int) ([]Procurement, []int, error) {
		res, err := c.GetProcurements(ctx, SearchByTimeParameters[ParameterTypeProcurements]{
			StartTime: startDate,
			EndTime:   endDate,
			Page:      page,
		})
		if err != nil {
			return nil, nil, err
		}

		return res.Procurements, res.OtherPages, nil
	})
}

func (c *Client) GetProcurements(ctx context.Context, params SearchByTimeParameters[ParameterTypeProcurements]) (ProcurementsResponse, error) {
	res, err := c.get(ctx, fmt.Sprintf("/laporan-transaksi-pembelian-obat-batch/index?%s", params.ToQuery(dateFormat)))
	if err != nil {
		return ProcurementsResponse{}, fmt.Errorf("get procurements with params %+v: %w", params, err)
	}
	defer res.Body.Close()

	procurements, err := ParseProcurements(res.Body)
	if err != nil {
		return ProcurementsResponse{}, fmt.Errorf("parse procurements with params %+v: %w", params, err)
	}

	return procurements, nil
}

func ParseProcurements(r io.Reader) (ProcurementsResponse, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return ProcurementsResponse{}, fmt.Errorf("new document from reader: %w", err)
	}

	procurementsSelections := doc.Find("div#w8-container > table > tbody > tr[data-key]")

	var procurements []Procurement
	procurementsSelections.EachWithBreak(func(i int, s *goquery.Selection) bool {
		procurement, parseErr := parseProcurement(s)
		if parseErr != nil {
			err = fmt.Errorf("parse procurement #%d: %w", i, parseErr)
			return false
		}

		procurements = append(procurements, procurement)
		return true
	})
	if err != nil {
		return ProcurementsResponse{}, err
	}

	return ProcurementsResponse{
		Procurements: procurements,
		OtherPages:   parsePagination(doc),
	}, nil
}

func parseProcurement(selection *goquery.Selection) (Procurement, error) {
	var procurement Procurement
	if err := UnmarshalDataColumn("procurement-column", selection, &procurement); err != nil {
		return Procurement{}, fmt.Errorf("unmarshal data column: %w", err)
	}

	procurementUnitsSelections := selection.Find("td[data-col-seq='0'] tr[data-key]")

	var err error
	procurementUnitsSelections.EachWithBreak(func(i int, s *goquery.Selection) bool {
		var procurementUnit ProcurementUnit
		if unmarshalErr := UnmarshalDataColumnByIndex("procurement-index", s, &procurementUnit); unmarshalErr != nil {
			err = fmt.Errorf("unmarshal procurement unit #%d: %w", i, unmarshalErr)
			return false
		}

		procurement.ProcurementUnits = append(procurement.ProcurementUnits, procurementUnit)
		return true
	})
	if err != nil {
		return Procurement{}, err
	}

	return procurement, nil
}
