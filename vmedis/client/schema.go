package client

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// DrugStock is the stock of a drug.
type DrugStock struct {
	Drug  Drug  `oos-column:"<self>"`
	Stock Stock `oos-column:"8"`
}

// Drug is a drug in the inventory.
type Drug struct {
	VmedisID     int
	VmedisCode   string `oos-column:"4" drugs-index:"3"`
	Name         string `oos-column:"5" drugs-index:"4"`
	Manufacturer string `oos-column:"12" drugs-index:"5"`
	Supplier     string `oos-column:"13"`
	MinimumStock Stock  `oos-column:"6"`
}

// Stock represents one instance of stock.
type Stock struct {
	Unit     string
	Quantity float64
}

// UnmarshalDataColumn implements DataColumnUnmarshaler.
func (s *Stock) UnmarshalDataColumn(selection *goquery.Selection) error {
	stockString := selection.Text()
	stockString = strings.TrimSpace(stockString)

	split := strings.Split(stockString, " ")
	if len(split) > 0 {
		// Here, the string can be either in "1.000,00" format or "1000.00" format.
		// We need to predict which format it is and convert it to float64.

		var qStr string

		// If there are always 3 digits after the dot, then it is in "1.000,00" format. Otherwise, it is in "1000.00" format.
		if strings.Count(split[0], ".") == 0 {
			qStr = strings.ReplaceAll(split[0], ",", ".")
		} else {
			firstFormat := true

			dotSplit := strings.Split(split[0], ".")
			for i := 1; i < len(dotSplit); i++ {
				beforeComma := strings.Split(dotSplit[i], ",")[0]
				if len(beforeComma) != 3 {
					firstFormat = false
					break
				}
			}

			if firstFormat {
				qStr = strings.ReplaceAll(split[0], ".", "")
				qStr = strings.ReplaceAll(qStr, ",", ".")
			} else {
				qStr = split[0]
			}
		}

		q, err := strconv.ParseFloat(qStr, 64)
		if err != nil {
			return fmt.Errorf("parse quantity from string [%s]: %w", split[0], err)
		}

		s.Quantity = q
	}

	if len(split) > 1 {
		s.Unit = split[1]
	}

	return nil
}
