package dumper

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
)

// DumpDailySales dumps the daily sales.
func DumpDailySales(ctx context.Context, db *gorm.DB, vmedisClient *client.Client) {
	log.Println("Dumping daily sales")

	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	sales, err := vmedisClient.GetAllTodaySales(ctx)
	if err != nil {
		log.Printf("Error getting daily sales: %s\n", err)
		return
	}

	salesModels := make([]models.Sale, len(sales))
	for i, sale := range sales {
		salesModels[i] = models.Sale{
			VmedisID:      sale.ID,
			SoldAt:        sale.Date.Time,
			InvoiceNumber: sale.InvoiceNumber,
			PatientName:   sale.PatientName,
			Doctor:        sale.Doctor,
			Payment:       sale.Payment,
			Total:         sale.Total,
			SaleUnits:     make([]models.SaleUnit, len(sale.SaleUnits)),
		}

		for j, saleUnit := range sale.SaleUnits {
			salesModels[i].SaleUnits[j] = models.SaleUnit{
				InvoiceNumber: sale.InvoiceNumber,
				IDInSale:      saleUnit.IDInSale,
				DrugCode:      saleUnit.DrugCode,
				DrugName:      saleUnit.DrugName,
				Batch:         saleUnit.Batch,
				Amount:        saleUnit.Amount,
				Unit:          saleUnit.Unit,
				UnitPrice:     saleUnit.UnitPrice,
				PriceCategory: saleUnit.PriceCategory,
				Discount:      saleUnit.Discount,
				Tuslah:        saleUnit.Tuslah,
				Embalase:      saleUnit.Embalase,
				Total:         saleUnit.Total,
			}
		}
	}

	if err := dumpSales(db, salesModels); err != nil {
		log.Printf("Error inserting daily sales: %s\n", err)
		return
	}
}

func dumpSales(db *gorm.DB, sales []models.Sale) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "invoice_number"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"updated_at",
				"vmedis_id",
				"sold_at",
				"patient_name",
				"doctor",
				"payment",
				"total",
			}),
		}).Omit("SaleUnits").Create(&sales).Error; err != nil {
			return fmt.Errorf("create sales: %w", err)
		}

		for _, sale := range sales {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "invoice_number"}, {Name: "id_in_sale"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"updated_at",
					"drug_code",
					"drug_name",
					"batch",
					"amount",
					"unit",
					"unit_price",
					"price_category",
					"discount",
					"tuslah",
					"embalase",
					"total",
				}),
			}).Create(&sale.SaleUnits).Error; err != nil {
				return fmt.Errorf("create sale units: %w", err)
			}
		}

		return nil
	})
}
