package procurement

import (
	"reflect"
	"testing"

	"github.com/turfaa/vmedis-proxy-api/drug"
)

func TestCalculateRecommendation(t *testing.T) {
	// ANTASIDA DOEN TAB 10x10 (MUTIFA): 168 strips in stock, 511 minimum.
	stock := drug.WithStock{
		Drug:  drug.Drug{VmedisCode: "OBT2312090001", MinimumStock: drug.Stock{Unit: "Strip", Quantity: 511}},
		Stock: drug.Stock{Unit: "Strip", Quantity: 168},
	}
	units := []drug.Unit{
		{Unit: "Strip", UnitOrder: 1},
		{Unit: "Box", ParentUnit: "Strip", ConversionToParentUnit: 10, UnitOrder: 2},
	}

	chosen, alternatives := calculateRecommendation(stock, units)

	// 511 * 2 - 168 = 854 strips, about 85 boxes.
	if want := (drug.Stock{Unit: "Box", Quantity: 85}); chosen != want {
		t.Errorf("got chosen %+v, want %+v", chosen, want)
	}
	if want := []drug.Stock{{Unit: "Strip", Quantity: 854}}; !reflect.DeepEqual(alternatives, want) {
		t.Errorf("got alternatives %+v, want %+v", alternatives, want)
	}
}
