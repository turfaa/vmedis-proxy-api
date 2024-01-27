package drug

import (
	"sort"
	"strings"
)

var (
	// minAllowedPriceByUnit defines the control that if the price of that unit is less than this value,
	// then the unit is ignored.
	//
	// If this value is 0, then the unit is always ignored.
	minAllowedPriceByUnit = map[string]float64{
		"tablet": 2000,
		"kapsul": 2000,
		"kaplet": 2000,

		"(jangan dipakai)": 0,
	}
)

func filterUnits(units []Unit) []Unit {
	sort.Slice(units, func(i, j int) bool {
		return units[i].UnitOrder < units[j].UnitOrder
	})

	result := make([]Unit, 0, len(units))
	conversion := 1.
	for _, u := range units {
		conversion *= u.ConversionToParentUnit

		minPrice, ok := minAllowedPriceByUnit[strings.ToLower(u.Unit)]
		if ok && u.PriceOne < minPrice {
			continue
		}

		var pu string
		if len(result) > 0 {
			pu = result[len(result)-1].Unit
		}

		result = append(result, Unit{
			Unit:                   u.Unit,
			ParentUnit:             pu,
			ConversionToParentUnit: conversion,
			UnitOrder:              len(result),
			PriceOne:               u.PriceOne,
			PriceTwo:               u.PriceTwo,
			PriceThree:             u.PriceThree,
		})
		conversion = 1
	}

	return result
}
