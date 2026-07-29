package database

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UndeleteAndUpdateColumns builds the assignments of an upsert's DO UPDATE
// clause for a soft-deletable model: it refreshes the given columns from the
// row being inserted, and clears deleted_at.
//
// The unique constraints of the soft-deletable models span their soft-deleted
// rows, so an upsert of a record that was soft-deleted here conflicts with the
// hidden row rather than inserting a second one. Clearing deleted_at is what
// makes that upsert bring the record back to life instead of quietly
// refreshing a row nothing can see.
func UndeleteAndUpdateColumns(columns []string) clause.Set {
	return append(
		clause.AssignmentColumns(columns),
		clause.Assignment{Column: clause.Column{Name: "deleted_at"}, Value: gorm.Expr("NULL")},
	)
}
