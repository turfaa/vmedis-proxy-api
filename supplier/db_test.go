package supplier

import (
	"context"
	"testing"

	"github.com/turfaa/vmedis-proxy-api/database"
	"github.com/turfaa/vmedis-proxy-api/database/models"
	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

// TestSupplierDumpRoundTrip checks that dumping upserts suppliers by their
// vmedis IDs, soft-deletes the ones no longer in vmedis, and revives a
// soft-deleted supplier when it shows up in vmedis again.
func TestSupplierDumpRoundTrip(t *testing.T) {
	ctx := context.Background()

	db, err := database.SqliteDB(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	d := NewDatabase(db)

	dump := func(suppliers ...vmedisv2.Supplier) int64 {
		t.Helper()

		if err := d.UpsertVmedisSuppliers(ctx, suppliers); err != nil {
			t.Fatalf("upsert suppliers: %v", err)
		}

		ids := make([]int, len(suppliers))
		for i, s := range suppliers {
			ids[i] = s.ID
		}

		deleted, err := d.SoftDeleteSuppliersNotIn(ctx, ids)
		if err != nil {
			t.Fatalf("soft-delete suppliers: %v", err)
		}
		return deleted
	}

	names := func() map[int]string {
		t.Helper()

		var suppliers []models.Supplier
		if err := db.Find(&suppliers).Error; err != nil {
			t.Fatalf("find suppliers: %v", err)
		}

		out := make(map[int]string, len(suppliers))
		for _, s := range suppliers {
			out[s.VmedisID] = s.Name
		}
		return out
	}

	a := vmedisv2.Supplier{ID: 1, Code: "SUP1", Name: "A", Active: true, Rating: 4.5}
	b := vmedisv2.Supplier{ID: 2, Code: "SUP2", Name: "B", Active: true}

	if deleted := dump(a, b); deleted != 0 {
		t.Errorf("first dump soft-deleted %d suppliers, want 0", deleted)
	}

	// B was removed from vmedis, and A was renamed.
	a.Name = "A renamed"
	if deleted := dump(a); deleted != 1 {
		t.Errorf("second dump soft-deleted %d suppliers, want 1", deleted)
	}
	if got := names(); len(got) != 1 || got[1] != "A renamed" {
		t.Errorf("after the second dump, got suppliers %v, want only A renamed", got)
	}

	// B is back in vmedis.
	if deleted := dump(a, b); deleted != 0 {
		t.Errorf("third dump soft-deleted %d suppliers, want 0", deleted)
	}
	if got := names(); len(got) != 2 || got[2] != "B" {
		t.Errorf("after the third dump, got suppliers %v, want A renamed and B", got)
	}

	var count int64
	if err := db.Unscoped().Model(&models.Supplier{}).Count(&count).Error; err != nil {
		t.Fatalf("count suppliers: %v", err)
	}
	if count != 2 {
		t.Errorf("got %d supplier rows including soft-deleted ones, want 2", count)
	}
}
