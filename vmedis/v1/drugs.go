package vmedisv1

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

// DrugsResponse is the response of the Drugs client method.
type DrugsResponse struct {
	Drugs      []Drug
	OtherPages []int
}

// GetAllDrugs gets all the drugs from vmedis.
// It fetches every /obat-batch/index?page=<page> page concurrently and
// returns an error if any page cannot be fetched or parsed.
func (c *Client) GetAllDrugs(ctx context.Context) ([]Drug, error) {
	return getAllPages(ctx, "drugs", c.concurrency, func(ctx context.Context, page int) ([]Drug, []int, error) {
		res, err := c.GetDrugs(ctx, page)
		if err != nil {
			return nil, nil, err
		}

		return res.Drugs, res.OtherPages, nil
	})
}

// GetDrugs gets the drugs from one page of "Data Obat" page in vmedis.
// It calls the /obat-batch/index?page=<page> page and try to parse the drugs from it.
func (c *Client) GetDrugs(ctx context.Context, page int) (DrugsResponse, error) {
	res, err := c.get(ctx, fmt.Sprintf("/obat-batch/index?page=%d", page))
	if err != nil {
		return DrugsResponse{}, fmt.Errorf("get drugs: %w", err)
	}
	defer res.Body.Close()

	drugs, err := ParseDrugs(res.Body)
	if err != nil {
		return DrugsResponse{}, fmt.Errorf("parse drugs: %w", err)
	}

	return drugs, nil
}

// ParseDrugs parses the drugs from the given reader.
func ParseDrugs(r io.Reader) (DrugsResponse, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return DrugsResponse{}, fmt.Errorf("parse HTML: %w", err)
	}

	var drugs []Drug
	doc.Find("tr[data-key]").EachWithBreak(func(i int, s *goquery.Selection) bool {
		drug, parseErr := parseDrug(s)
		if parseErr != nil {
			err = fmt.Errorf("parse drug #%d: %w", i, parseErr)
			return false
		}

		drugs = append(drugs, drug)
		return true
	})
	if err != nil {
		return DrugsResponse{}, err
	}

	return DrugsResponse{Drugs: drugs, OtherPages: parsePagination(doc)}, nil
}

func parseDrug(selection *goquery.Selection) (Drug, error) {
	var drug Drug
	if err := UnmarshalDataColumnByIndex("drugs-index", selection, &drug); err != nil {
		return Drug{}, fmt.Errorf("unmarshal drug: %w", err)
	}

	// Get the value from <a class="pilih" value="123" href="/obat-batch/index" title="Detail">some image</a>
	idStr, ok := selection.Find("a.pilih").Attr("value")
	if !ok {
		return Drug{}, fmt.Errorf("drug vmedis id not found")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return Drug{}, fmt.Errorf("parse drug vmedis id: %w", err)
	}

	drug.VmedisID = id

	return drug, nil
}
