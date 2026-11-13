package procurement_test

import (
	"encoding/json"
	jsonv2 "encoding/json/v2"
	"testing"

	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/procurement"
)

// TestRecommendationJSONEmbedsDrugStock checks that a recommendation's drug and
// stock are flattened into the recommendation object, both through the v1 Redis
// round-trip and the v2 HTTP renderer. The dashboard reads `drug` and `stock`
// at the top level.
func TestRecommendationJSONEmbedsDrugStock(t *testing.T) {
	want := procurement.Recommendation{
		DrugStock: drug.WithStock{
			Drug:  drug.Drug{VmedisCode: "ABC"},
			Stock: drug.Stock{Unit: "Box", Quantity: 1},
		},
		Procurement: drug.Stock{Unit: "Box", Quantity: 2},
	}

	stored, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("v1 marshal: %v", err)
	}

	var got procurement.Recommendation
	if err := json.Unmarshal(stored, &got); err != nil {
		t.Fatalf("v1 unmarshal: %v", err)
	}
	if got.DrugStock.Drug.VmedisCode != "ABC" {
		t.Fatalf("v1 round-trip lost the drug: %s", stored)
	}

	served, err := jsonv2.Marshal(got)
	if err != nil {
		t.Fatalf("v2 marshal: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(served, &fields); err != nil {
		t.Fatalf("unmarshal served JSON: %v", err)
	}
	for _, key := range []string{"drug", "stock", "procurement"} {
		if _, ok := fields[key]; !ok {
			t.Errorf("served JSON is missing top-level %q: %s", key, served)
		}
	}
	if _, ok := fields["DrugStock"]; ok {
		t.Errorf("served JSON nests the drug under DrugStock: %s", served)
	}
}
