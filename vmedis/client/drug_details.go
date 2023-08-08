package client

import (
	"context"
	"fmt"
	"io"

	"github.com/PuerkitoBio/goquery"
)

// GetDrug gets the drug details from vmedis.
// It calls the /obat-batch/view?id=<id> page and try to parse the drug from it.
func (c *Client) GetDrug(ctx context.Context, id int) (Drug, error) {
	res, err := c.get(ctx, fmt.Sprintf("/obat-batch/view?id=%d", id))
	if err != nil {
		return Drug{}, fmt.Errorf("get drug: %w", err)
	}
	defer res.Body.Close()

	drug, err := ParseDrugDetails(res.Body)
	if err != nil {
		return Drug{}, fmt.Errorf("parse drug: %w", err)
	}

	drug.VmedisID = id
	return drug, nil
}

// ParseDrugDetails parses the drug from the given reader
// It usually comes from the /obat-batch/view?id=<id> page.
func ParseDrugDetails(r io.Reader) (Drug, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return Drug{}, fmt.Errorf("parse HTML: %w", err)
	}

	var drug Drug
	if err := UnmarshalForm(doc.Selection, &drug); err != nil {
		return Drug{}, fmt.Errorf("unmarshal drug details: %w", err)
	}

	units, err := EnrichUnitsFromDoc(drug.Units, doc)
	if err != nil {
		return Drug{}, fmt.Errorf("enrich units of drug %s: %w", drug.Name, err)
	}

	drug.Units = units
	if len(units) > 0 {
		drug.MinimumStock.Unit = units[0].Unit
	}

	return drug, nil
}
