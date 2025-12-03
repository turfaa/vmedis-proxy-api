package vmedis

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ProcurementsResponse struct {
	Procurements []Procurement
	OtherPages   []int
}

func (c *Client) GetAllProcurementsBetweenDates(
	ctx context.Context,
	startDate time.Time,
	endDate time.Time,
) ([]Procurement, error) {
	var (
		procurements []Procurement
		pages        = make(chan int, c.concurrency*2)
		wg           sync.WaitGroup
		lock         sync.Mutex
	)

	// Get the number of pages
	log.Printf("Getting number of pages of procurements between %s and %s\n", startDate.Format(dateFormat), endDate.Format(dateFormat))
	res, err := c.GetProcurements(ctx, SearchByTimeParameters[ParameterTypeProcurements]{
		StartTime: startDate,
		EndTime:   endDate,
		Page:      9999999,
	})
	if err != nil {
		return nil, fmt.Errorf("get number of pages: %w", err)
	}

	lastPage := 1
	for _, p := range res.OtherPages {
		if p > lastPage {
			lastPage = p
		}
	}

	log.Printf("Number of procurement pages: %d\n", lastPage)

	go func() {
		for i := 1; i <= lastPage; i++ {
			pages <- i
		}
		close(pages)
	}()

	// Start the workers
	for i := 0; i < c.concurrency; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for page := range pages {
				log.Printf("Getting procurements at page %d\n", page)

				res, err := c.GetProcurements(ctx, SearchByTimeParameters[ParameterTypeProcurements]{
					StartTime: startDate,
					EndTime:   endDate,
					Page:      page,
				})
				if err != nil {
					log.Printf("Error getting procurements at page #%d: %v\n", page, err)
					continue
				}

				lock.Lock()
				procurements = append(procurements, res.Procurements...)
				lock.Unlock()

				log.Printf("Got %d procurements at page %d\n", len(res.Procurements), page)
			}
		}()
	}

	wg.Wait()

	return procurements, nil
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

	procurementsSelections := doc.Find("div#w5-container > table > tbody > tr[data-key]")

	var procurements []Procurement
	procurementsSelections.Each(func(i int, s *goquery.Selection) {
		procurement, err := parseProcurement(s)
		if err != nil {
			log.Printf("error parsing procurement #%d: %s", i, err)
			return
		}
		procurements = append(procurements, procurement)
	})

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

	procurementUnitsSelections.Each(func(i int, s *goquery.Selection) {
		var procurementUnit ProcurementUnit
		if err := UnmarshalDataColumnByIndex("procurement-index", s, &procurementUnit); err != nil {
			log.Printf("error unmarshaling procurement unit #%d: %s", i, err)
			return
		}

		procurement.ProcurementUnits = append(procurement.ProcurementUnits, procurementUnit)
	})

	return procurement, nil
}
