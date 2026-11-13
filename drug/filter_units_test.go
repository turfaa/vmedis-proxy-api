package drug

import (
	"reflect"
	"testing"
)

func TestFilterUnits(t *testing.T) {
	tests := []struct {
		name  string
		units []Unit
		want  []Unit
	}{
		{
			// The tablet is too cheap to be shown, so the strip is the first
			// unit shown, without a conversion, rather than "Strip (10 )".
			name: "hidden smallest unit",
			units: []Unit{
				{Unit: "Tablet", UnitOrder: 0, PriceOne: 750},
				{Unit: "Strip", ParentUnit: "Tablet", ConversionToParentUnit: 10, UnitOrder: 1, PriceOne: 7500},
				{Unit: "Box", ParentUnit: "Strip", ConversionToParentUnit: 10, UnitOrder: 2, PriceOne: 62000},
			},
			want: []Unit{
				{Unit: "Strip", UnitOrder: 0, PriceOne: 7500},
				{Unit: "Box", ParentUnit: "Strip", ConversionToParentUnit: 10, UnitOrder: 1, PriceOne: 62000},
			},
		},
		{
			// Stored with a conversion of 1 for the smallest unit, as vmedis
			// gives it, rather than "Strip (1 )".
			name: "smallest unit with a conversion",
			units: []Unit{
				{Unit: "Strip", ConversionToParentUnit: 1, UnitOrder: 0, PriceOne: 7500},
				{Unit: "Box", ParentUnit: "Strip", ConversionToParentUnit: 10, UnitOrder: 1, PriceOne: 60500},
			},
			want: []Unit{
				{Unit: "Strip", UnitOrder: 0, PriceOne: 7500},
				{Unit: "Box", ParentUnit: "Strip", ConversionToParentUnit: 10, UnitOrder: 1, PriceOne: 60500},
			},
		},
		{
			// A hidden unit in the middle folds its conversion into the next one.
			name: "hidden middle unit",
			units: []Unit{
				{Unit: "Botol", UnitOrder: 0, PriceOne: 10000},
				{Unit: "(JANGAN DIPAKAI)", ParentUnit: "Botol", ConversionToParentUnit: 6, UnitOrder: 1, PriceOne: 55000},
				{Unit: "Dus", ParentUnit: "(JANGAN DIPAKAI)", ConversionToParentUnit: 4, UnitOrder: 2, PriceOne: 200000},
			},
			want: []Unit{
				{Unit: "Botol", UnitOrder: 0, PriceOne: 10000},
				{Unit: "Dus", ParentUnit: "Botol", ConversionToParentUnit: 24, UnitOrder: 1, PriceOne: 200000},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterUnits(tt.units); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got\n%+v\nwant\n%+v", got, tt.want)
			}
		})
	}
}
