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
	VmedisCode   string `oos-column:"4" drugs-index:"3" form-name:"Obat[obatkode]"`
	Name         string `oos-column:"5" drugs-index:"4" form-name:"Obat[obatnama]"`
	Manufacturer string `oos-column:"12" drugs-index:"5" form-name:"Obat[pabid]"`
	Supplier     string `oos-column:"13"`
	MinimumStock Stock  `oos-column:"6" form-name:"Obat[obatminstok]"`
	Units        []Unit `form-name:"Obat[soid%d]"`
}

// Stock represents one instance of stock.
type Stock struct {
	Unit     string
	Quantity float64
}

// UnmarshalDataColumn implements DataColumnUnmarshaler.
func (s *Stock) UnmarshalDataColumn(selection *goquery.Selection) error {
	stockString := selection.Text()
	return s.UnmarshalText([]byte(stockString))
}

// UnmarshalForm implements FormUnmarshaler.
func (s *Stock) UnmarshalForm(selection *goquery.Selection) error {
	stockString := selection.AttrOr("value", "")
	return s.UnmarshalText([]byte(stockString))
}

// UnmarshalText implements TextUnmarshaler.
func (s *Stock) UnmarshalText(text []byte) error {
	stockString := string(text)
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

		var q float64
		if qStr != "" {
			var err error
			q, err = strconv.ParseFloat(qStr, 64)
			if err != nil {
				return fmt.Errorf("parse quantity from string [%s]: %w", split[0], err)
			}
		}

		s.Quantity = q
	}

	if len(split) > 1 {
		s.Unit = split[1]
	}

	return nil
}

// Unit represents a unit of a drug.
type Unit struct {
	Unit                   string
	ParentUnit             string
	ConversionToParentUnit float64

	// UnitOrder is the order of the unit of the drug.
	// The smallest unit has the lowest order.
	UnitOrder int

	formName string
}

// UnmarshalForm implements FormUnmarshaler.
// It only parses the unit name.
// To complete the unit data, use the EnrichUnitsFromDoc function.
func (u *Unit) UnmarshalForm(selection *goquery.Selection) error {
	u.formName = selection.AttrOr("name", "")
	u.Unit = selection.AttrOr("value", "")
	return nil
}

// EnrichUnitsFromDoc completes the units data by using the data in the goquery document.
// It also filters out the units that doesn't have a form name or a name.
// The document usually comes from the /obat-batch/view?id=<id> page.
func EnrichUnitsFromDoc(units []Unit, doc *goquery.Document) ([]Unit, error) {
	result := make([]Unit, 0, len(units))

	for _, u := range units {
		if u.formName == "" || u.Unit == "" {
			continue
		}

		selection := doc.Selection

		unitNameSelection := selection.Find("input[name=\"" + u.formName + "\"]")
		if unitNameSelection.Length() != 0 {
			u.Unit = unitNameSelection.AttrOr("value", "")
		}

		if len(result) > 0 {
			u.ParentUnit = result[len(result)-1].Unit
		}

		conversionSelection := selection.Find("input[name=\"" + strings.ReplaceAll(u.formName, "soid", "sodkonversi") + "\"]")
		if conversionSelection.Length() > 0 {
			conversionStr := conversionSelection.AttrOr("value", "")
			conversion, err := strconv.ParseFloat(conversionStr, 64)
			if err != nil {
				html, _ := conversionSelection.Html()
				return nil, fmt.Errorf("parse '%s' conversion from string [%s] <%s>: %w", u.Unit, conversionStr, html, err)
			}

			u.ConversionToParentUnit = conversion
		}

		u.UnitOrder = len(result)
	
		result = append(result, u)
	}

	return result, nil
}
