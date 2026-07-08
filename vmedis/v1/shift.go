package vmedisv1

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ShiftsResponse struct {
	Shifts     []Shift
	OtherPages []int
}

// GetAllShiftsBetweenTimes gets all the shifts between the given times from vmedis.
// It fetches every /laporan-gantishift/index page concurrently and returns an
// error if any page cannot be fetched or parsed.
func (c *Client) GetAllShiftsBetweenTimes(ctx context.Context, startTime time.Time, endTime time.Time) ([]Shift, error) {
	return getAllPages(ctx, "shifts", c.concurrency, func(ctx context.Context, page int) ([]Shift, []int, error) {
		res, err := c.GetShifts(ctx, SearchByTimeParameters[ParameterTypeShifts]{
			StartTime: startTime,
			EndTime:   endTime,
			Page:      page,
		})
		if err != nil {
			return nil, nil, err
		}

		return res.Shifts, res.OtherPages, nil
	})
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
	doc.Find("tr[data-key]").EachWithBreak(func(i int, s *goquery.Selection) bool {
		shift, parseErr := parseShift(s)
		if parseErr != nil {
			err = fmt.Errorf("parse shift #%d: %w", i, parseErr)
			return false
		}

		shifts = append(shifts, shift)
		return true
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
