package drug

import (
	"context"
	"reflect"
	"testing"

	"github.com/turfaa/vmedis-proxy-api/database"
	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

// TestUpsertVmedisDrugUnitsDeletesUnitsNoLongerInVmedis checks that a unit
// renamed in vmedis does not stay behind as a second smallest unit, which made
// the procurement recommendation of ANTASIDA DOEN TAB use its old "BLISTER"
// unit instead of "Box".
func TestUpsertVmedisDrugUnitsDeletesUnitsNoLongerInVmedis(t *testing.T) {
	ctx := context.Background()

	db, err := database.SqliteDB(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	d := NewDatabase(db)

	upsert := func(code string, units []vmedisv2.Unit) {
		t.Helper()
		if err := d.UpsertVmedisDrugUnits(ctx, code, units); err != nil {
			t.Fatalf("upsert units of %s: %v", code, err)
		}
	}

	// The units before the rename, and another drug's units.
	upsert("ANTASIDA", []vmedisv2.Unit{
		{Unit: "BLISTER", UnitOrder: 1, PriceOne: 2000},
	})
	upsert("OTHER", []vmedisv2.Unit{
		{Unit: "Tablet", UnitOrder: 1},
	})

	// vmedis now has "Strip" and "Box" instead of "BLISTER".
	upsert("ANTASIDA", []vmedisv2.Unit{
		{Unit: "Strip", UnitOrder: 1, PriceOne: 2000},
		{Unit: "Box", ParentUnit: "Strip", ConversionToParentUnit: 10, UnitOrder: 2, PriceOne: 13000},
	})

	// Reading units that could not be parsed must not delete everything.
	upsert("OTHER", nil)

	got, err := d.GetDrugUnitsByDrugVmedisCodes(ctx, []string{"ANTASIDA", "OTHER"})
	if err != nil {
		t.Fatalf("get units: %v", err)
	}

	unitNames := func(units []Unit) []string {
		names := make([]string, len(units))
		for i, u := range units {
			names[i] = u.Unit
		}
		return names
	}

	if want := []string{"Strip", "Box"}; !reflect.DeepEqual(unitNames(got["ANTASIDA"]), want) {
		t.Errorf("got ANTASIDA units %v, want %v", unitNames(got["ANTASIDA"]), want)
	}
	if want := []string{"Tablet"}; !reflect.DeepEqual(unitNames(got["OTHER"]), want) {
		t.Errorf("got OTHER units %v, want %v", unitNames(got["OTHER"]), want)
	}
}
