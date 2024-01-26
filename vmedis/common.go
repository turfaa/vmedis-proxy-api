package vmedis

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parsePagination(doc *goquery.Document) []int {
	var otherPages []int
	doc.Find(".pagination li a").Each(func(i int, s *goquery.Selection) {
		page, err := strconv.Atoi(s.Text())
		if err != nil {
			// expected, ignore
			return
		}

		otherPages = append(otherPages, page)
	})

	return otherPages
}

func parseFloat(s string) (float64, error) {
	// Here, the string can be either in "1.000,00" format or "1000.00" format.
	// We need to predict which format it is and convert it to float64.

	var fStr string

	// If there are always 3 digits after the dot, then it is in "1.000,00" format. Otherwise, it is in "1000.00" format.
	if strings.Count(s, ".") == 0 {
		fStr = strings.ReplaceAll(s, ",", ".")
	} else {
		firstFormat := true

		dotSplit := strings.Split(s, ".")
		for i := 1; i < len(dotSplit); i++ {
			beforeComma := strings.Split(dotSplit[i], ",")[0]
			if len(beforeComma) != 3 {
				firstFormat = false
				break
			}
		}

		if firstFormat {
			fStr = strings.ReplaceAll(s, ".", "")
			fStr = strings.ReplaceAll(fStr, ",", ".")
		} else {
			fStr = s
		}
	}

	var f float64
	if fStr != "" {
		var err error
		f, err = strconv.ParseFloat(fStr, 64)
		if err != nil {
			return 0, fmt.Errorf("parse float from string [%s]: %w", s, err)
		}
	}

	return f, nil
}
