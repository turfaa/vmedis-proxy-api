package database

import (
	"fmt"

	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PrepopulateInvoiceCalculators prepopulates invoice calculators.
func PrepopulateInvoiceCalculators(db *gorm.DB) error {
	calculators := []models.InvoiceCalculator{
		{
			Supplier:    "Kuda Mas",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Kuda Mas",
					Name:       "Total",
					Multiplier: 0.99,
				},
			},
		},
		{
			Supplier:    "Nara Artha",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Nara Artha",
					Name:       "Total",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "Kwatro Mandiri Ekavisi",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Kwatro Mandiri Ekavisi",
					Name:       "Nilai Dokumen",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "Anugerah Pharmindo Lestari",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Anugerah Pharmindo Lestari",
					Name:       "Jumlah Tagihan",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "Bintang Mega Mandiri",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Bintang Mega Mandiri",
					Name:       "Grand Total",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "Rubel Anugerah Medicatama",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Rubel Anugerah Medicatama",
					Name:       "SubJumlah/DPP",
					Multiplier: 0.99,
				},
				{
					Supplier:   "Rubel Anugerah Medicatama",
					Name:       "PPn",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "Hasil Karya Sejahtera",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Hasil Karya Sejahtera",
					Name:       "Total",
					Multiplier: 0.985,
				},
			},
		},
		{
			Supplier:    "Surya Donasin",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Surya Donasin",
					Name:       "Jumlah Tagihan",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "Masamedi Intifarmindo",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Masamedi Intifarmindo",
					Name:       "Jumlah Tagihan",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "Tempo",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Tempo",
					Name:       "Jumlah Tagihan",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "Triputra Mulia Farmasindo",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Triputra Mulia Farmasindo",
					Name:       "Total",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "Carmella Gustavindo",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Carmella Gustavindo",
					Name:       "Total",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "Jaya Bakti Raharja",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Jaya Bakti Raharja",
					Name:       "Jumlah Tagihan",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "Penta Valent",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Penta Valent",
					Name:       "Jumlah Tagihan",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "Dempo Sentosa",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Dempo Sentosa",
					Name:       "Grand Total",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "Sapta Sari Tama",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Sapta Sari Tama",
					Name:       "Jumlah Tagihan",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "San Prima Sejati",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "San Prima Sejati",
					Name:       "Total",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "Glory Majesty Indonesia",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Glory Majesty Indonesia",
					Name:       "Total",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "Bina San Prima",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Bina San Prima",
					Name:       "Harus Dibayar",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "Andalus Perta Adia",
			ShouldRound: true,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Andalus Perta Adia",
					Name:       "Total",
					Multiplier: 1,
				},
			},
		},
		{
			Supplier:    "Teknologi Medika Pratama",
			ShouldRound: false,
			Components: []models.InvoiceComponent{
				{
					Supplier:   "Teknologi Medika Pratama",
					Name:       "Total Tagihan",
					Multiplier: 1,
				},
			},
		},
	}

	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "supplier"}},
		UpdateAll: true,
	}).Omit(clause.Associations).Create(&calculators).Error; err != nil {
		return fmt.Errorf("create invoice calculators: %w", err)
	}

	for _, calculator := range calculators {
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "supplier"}, {Name: "name"}},
			UpdateAll: true,
		}).Create(calculator.Components).Error; err != nil {
			return fmt.Errorf("create invoice components: %w", err)
		}
	}

	return nil
}
