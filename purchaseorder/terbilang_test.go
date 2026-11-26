package purchaseorder

import (
	"encoding/json"
	"os"
	"testing"
)

// The cases are in testdata/terbilang.json, so the dashboard's copy of
// terbilang can be checked against the same ones.
func TestTerbilang(t *testing.T) {
	raw, err := os.ReadFile("testdata/terbilang.json")
	if err != nil {
		t.Fatalf("read cases: %v", err)
	}

	var cases [][2]any
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("parse cases: %v", err)
	}

	for _, c := range cases {
		n, want := uint64(c[0].(float64)), c[1].(string)
		if got := terbilang(n); got != want {
			t.Errorf("terbilang(%d) = %q, want %q", n, got, want)
		}
	}
}
