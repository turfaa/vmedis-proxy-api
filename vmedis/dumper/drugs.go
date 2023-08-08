package dumper

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
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

	var counter, errCounter atomic.Int32

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for drug := range drugs {
				log.Printf("Dumping drug: %s [%s]\n", drug.Name, drug.VmedisCode)

				drugM := models.Drug{
					VmedisID:     drug.VmedisID,
					VmedisCode:   drug.VmedisCode,
					Name:         drug.Name,
					Manufacturer: drug.Manufacturer,
				}

				if err := db.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "vmedis_code"}},
					DoUpdates: clause.AssignmentColumns([]string{"updated_at", "vmedis_id", "name", "manufacturer"}),
				}).Create(&drugM).Error; err != nil {
					log.Printf("Error creating drug: %s\n", err)
					errCounter.Add(1)
				} else {
					log.Printf("Dumped drug: %s [%s]\n", drug.Name, drug.VmedisCode)
					counter.Add(1)
				}
			}
		}()
	}

	wg.Wait()
	log.Printf("Finished dumping drugs: %d, errors: %d\n", counter.Load(), errCounter.Load())
}
