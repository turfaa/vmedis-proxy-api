package vmedis

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Percentage contains the percentage value.
// For example, if the value is 10%, then Percentage is 10.
type Percentage struct {
	Value float64
}

// UnmarshalDataColumn parses a string with "12,34 %" format.
func (p *Percentage) UnmarshalDataColumn(selection *goquery.Selection) error {
	percentageString := selection.Text()

	percentageString = strings.TrimSuffix(percentageString, "%")
	percentageString = strings.TrimSpace(percentageString)

	value, err := parseFloat(percentageString)
	if err != nil {
		return fmt.Errorf("parse percentage from string [%s]: %w", percentageString, err)
	}

	p.Value = value
	return nil
}
