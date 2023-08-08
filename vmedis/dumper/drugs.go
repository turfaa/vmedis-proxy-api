package dumper

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
)

const (
	drugsBatchSize = 20
)

// DumpDrugs dumps the drugs.
func DumpDrugs(db *gorm.DB, vmedisClient *client.Client) {
	log.Println("Dumping drugs")

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	defer cancel()

	drugs, err := vmedisClient.GetAllDrugs(ctx)
	if err != nil {
		log.Printf("Error getting drugs: %s\n", err)
		return
	}

	var (
		toInsert   []models.Drug
		counter    int
		errCounter int
	)
	for drug := range drugs {
		toInsert = append(toInsert, models.Drug{
			VmedisID:     drug.VmedisID,
			VmedisCode:   drug.VmedisCode,
			Name:         drug.Name,
			Manufacturer: drug.Manufacturer,
		})

		if len(toInsert) == drugsBatchSize {
			log.Printf("Dumping %d drugs\n", len(toInsert))
			if err := dumpDrugs(db, toInsert); err != nil {
				log.Printf("Error inserting drugs: %s\n", err)
				errCounter += len(toInsert)
			} else {
				toInsert = nil
				counter += len(toInsert)
			}
		}
	}

	if len(toInsert) > 0 {
		log.Printf("Dumping %d drugs\n", len(toInsert))
		if err := dumpDrugs(db, toInsert); err != nil {
			log.Printf("Error inserting drugs: %s\n", err)
			errCounter += len(toInsert)
		} else {
			counter += len(toInsert)
		}
	}

	log.Printf("Finished dumping drugs: %d, errors: %d\n", counter, errCounter)
}

func dumpDrugs(db *gorm.DB, drugs []models.Drug) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "vmedis_code"}},
		DoUpdates: clause.AssignmentColumns([]string{"updated_at", "vmedis_id", "name", "manufacturer"}),
	}).
		Create(&drugs).
		Error
}
