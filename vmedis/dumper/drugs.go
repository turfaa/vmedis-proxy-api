package dumper

import (
	"context"
	"log"
	"sync"
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
func DumpDrugs(ctx context.Context, db *gorm.DB, vmedisClient *client.Client, detailsPuller chan<- models.Drug) {
	log.Println("Dumping drugs")

	ctx, cancel := context.WithTimeout(ctx, 6*time.Hour)
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

		if len(toInsert) >= drugsBatchSize {
			log.Printf("Dumping %d drugs\n", len(toInsert))
			if err := dumpDrugs(db, toInsert); err != nil {
				log.Printf("Error inserting drugs: %s\n", err)
				errCounter += len(toInsert)
			} else {
				log.Printf("Inserted %d drugs\n", len(toInsert))
				counter += len(toInsert)
			}

			for _, d := range toInsert {
				detailsPuller <- d
			}

			toInsert = nil
		}
	}

	if len(toInsert) > 0 {
		log.Printf("Dumping %d drugs\n", len(toInsert))
		if err := dumpDrugs(db, toInsert); err != nil {
			log.Printf("Error inserting drugs: %s\n", err)
			errCounter += len(toInsert)
		} else {
			log.Printf("Inserted %d drugs\n", len(toInsert))
			counter += len(toInsert)
		}
	}

	log.Printf("Finished dumping drugs: %d, errors: %d\n", counter, errCounter)
}

// DrugDetailsPuller pulls the details of the drugs from the channel and store them into the database.
func DrugDetailsPuller(ctx context.Context, db *gorm.DB, vmedisClient *client.Client) (chan<- models.Drug, func()) {
	drugs := make(chan models.Drug, 100)
	closeChan := make(chan struct{})

	var wg sync.WaitGroup
	closeFunc := func() {
		close(closeChan)
		wg.Wait()
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case drug := <-drugs:
					log.Printf("Getting drug details of id %d\n", drug.VmedisID)

					ctx, cancel := context.WithTimeout(ctx, time.Minute)
					d, err := vmedisClient.GetDrug(ctx, drug.VmedisID)
					cancel()

					if err != nil {
						log.Printf("Error getting drug details of id %d: %s\n", drug.VmedisID, err)
						continue
					}

					drug.VmedisCode = d.VmedisCode
					drug.Name = d.Name
					drug.Manufacturer = d.Manufacturer
					drug.MinimumStock = models.Stock{
						Unit:     d.MinimumStock.Unit,
						Quantity: d.MinimumStock.Quantity,
					}

					log.Printf("Dumping drug details of id %d [%s]\n", drug.VmedisID, drug.Name)
					if err := dumpDrugDetails(db, drug); err != nil {
						log.Printf("Error inserting drug details of id %d: %s\n", drug.VmedisID, err)
					}

					var units []models.DrugUnit
					for _, u := range d.Units {
						units = append(units, models.DrugUnit{
							Unit:                   u.Unit,
							DrugVmedisCode:         d.VmedisCode,
							ParentUnit:             u.ParentUnit,
							ConversionToParentUnit: u.ConversionToParentUnit,
							UnitOrder:              u.UnitOrder,
							PriceOne:               u.PriceOne,
							PriceTwo:               u.PriceTwo,
							PriceThree:             u.PriceThree,
						})
					}

					if err := dumpDrugUnits(db, units); err != nil {
						log.Printf("Error inserting drug units of id %d: %s\n", drug.VmedisID, err)
					}

				case <-closeChan:
					return

				case <-ctx.Done():
					return
				}
			}
		}()
	}

	return drugs, closeFunc
}

func dumpDrugs(db *gorm.DB, drugs []models.Drug) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "vmedis_code"}},
		DoUpdates: clause.AssignmentColumns([]string{"updated_at", "vmedis_id", "name", "manufacturer"}),
	}).
		Create(&drugs).
		Error
}

func dumpDrugDetails(db *gorm.DB, drug models.Drug) error {
	columns := []string{"updated_at"}
	if drug.VmedisCode != "" {
		columns = append(columns, "vmedis_code")
	}

	if drug.Name != "" {
		columns = append(columns, "name")
	}

	if drug.Manufacturer != "" {
		columns = append(columns, "manufacturer")
	}

	if drug.MinimumStock.Unit != "" {
		columns = append(columns, "minimum_stock_unit", "minimum_stock_quantity")
	}

	if len(columns) > 1 {
		return db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "vmedis_id"}},
			DoUpdates: clause.AssignmentColumns(columns),
		}).
			Create(&drug).
			Error
	} else {
		log.Printf("No columns to update for drug %d\n", drug.VmedisID)
		return nil
	}
}

func dumpDrugUnits(db *gorm.DB, units []models.DrugUnit) error {
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "drug_vmedis_code"}, {Name: "unit"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"updated_at",
			"parent_unit",
			"conversion_to_parent_unit",
			"unit_order",
			"price_one",
			"price_two",
			"price_three",
		}),
	}).
		Create(&units).
		Error
}
