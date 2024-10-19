package vmedis

import (
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
)

type ShiftsResponse struct {
	Shifts     []Shift
	OtherPages []int
}

func (c *Client) GetAllShiftsBetweenTimes(ctx context.Context, startTime time.Time, endTime time.Time) ([]Shift, error) {
	log.Println("Getting number of pages of shifts")
	res, err := c.GetShifts(ctx, SearchByTimeParameters[ParameterTypeShifts]{
		StartTime: startTime,
		EndTime:   endTime,
		Page:      999999,
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

	log.Printf("Number of shifts pages: %d\n", lastPage)

	var (
		shifts []Shift
		pages  = make(chan int, c.concurrency*2)
		lock   sync.Mutex
	)

	go func() {
		for i := 1; i <= lastPage; i++ {
			pages <- i
		}
		close(pages)
	}()

	eg, ctx := errgroup.WithContext(ctx)

	for i := 0; i < c.concurrency; i++ {
		eg.Go(func() error {
			for page := range pages {
				log.Printf("Getting shifts at page %d\n", page)

				res, err := c.GetShifts(ctx, SearchByTimeParameters[ParameterTypeShifts]{
					StartTime: startTime,
					EndTime:   endTime,
					Page:      page,
				})
				if err != nil {
					return fmt.Errorf("get shifts at page %d: %w", page, err)
				}

				lock.Lock()
				shifts = append(shifts, res.Shifts...)
				lock.Unlock()
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("get all shifts: %w", err)
	}

	return shifts, nil
}

func (c *Client) GetShifts(ctx context.Context, params SearchByTimeParameters[ParameterTypeShifts]) (ShiftsResponse, error) {
	res, err := c.get(ctx, fmt.Sprintf("/laporan-gantishift/index?%s", params.ToQuery(dateTimeMinuteFormat)))
	if err != nil {
		return ShiftsResponse{}, fmt.Errorf("get shifts with params %+v: %w", params, err)
	}
	defer res.Body.Close()

	shifts, err := ParseShifts(res.Body)
	if err != nil {
		return ShiftsResponse{}, fmt.Errorf("parse shifts with params %+v: %w", params, err)
	}

	return shifts, nil
}

func ParseShifts(r io.Reader) (ShiftsResponse, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return ShiftsResponse{}, fmt.Errorf("parse HTML: %w", err)
	}

	var shifts []Shift
	doc.Find("tr[data-key]").Each(func(i int, s *goquery.Selection) {
		shift, innerErr := parseShift(s)
		if innerErr != nil {
			err = fmt.Errorf("parse shift #%d: %w", i, innerErr)
			return
		}

		shifts = append(shifts, shift)
	})
	if err != nil {
		return ShiftsResponse{}, err
	}

	return ShiftsResponse{Shifts: shifts, OtherPages: parsePagination(doc)}, nil
}

func parseShift(s *goquery.Selection) (Shift, error) {
	var shift Shift
	if err := UnmarshalDataColumnByIndex("shift-index", s, &shift); err != nil {
		return Shift{}, fmt.Errorf("parse shift: %w", err)
	}

	// Get the value from <button type="button" class="btn btn-warning btn-xs actionPrint" value="110844" title="Cetak Faktur">.
	idStr, ok := s.Find("button.actionPrint").Attr("value")
	if ok {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return Shift{}, fmt.Errorf("parse shift id: %w", err)
		}

		shift.ID = id
	} else {
		html, _ := s.Html()
		return Shift{}, fmt.Errorf("shift id not found in: %s", html)
	}

	return shift, nil
}
