package dumper

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis/proxy"
)

const (
	procurementRecommendationsKey = "static_key.procurement_recommendations.json.zlib"
)

// DumpProcurementRecommendations calculates and dumps procurement recommendations to cache.
func DumpProcurementRecommendations(ctx context.Context, db *gorm.DB, redisClient *redis.Client, vmedisClient *client.Client) {
	log.Println("Computing procurement recommendations and writing them to cache....")

	ctx, cancel := context.WithTimeout(ctx, time.Minute*30)
	defer cancel()

	log.Printf("Getting all out-of-stock-drugs for writing procurement recommendations")
	oosDrugs, err := vmedisClient.GetAllOutOfStockDrugs(ctx)
	if err != nil {
		log.Printf("Error getting all out of stock drugs for writing procurement recommendations: %s\n", err)
		return
	}

	log.Printf("Got %d out-of-stock drugs for writing procurement recommendations\n", len(oosDrugs))

	log.Printf("Getting drug units of out-of-stock drugs")
	drugCodes := make([]string, len(oosDrugs))
	for i, drug := range oosDrugs {
		drugCodes[i] = drug.Drug.VmedisCode
	}

	unitsByCode, err := getDrugUnits(db, drugCodes)
	if err != nil {
		log.Printf("Error getting drug units of out-of-stock drugs: %s\n", err)
		return
	}
	log.Printf("Got %d drug units of out-of-stock drugs\n", len(unitsByCode))

	recommendations := make([]proxy.DrugProcurementRecommendation, len(oosDrugs))
	for i, drugStock := range oosDrugs {
		recommendations[i] = proxy.DrugProcurementRecommendation{
			DrugStock:    proxy.FromClientDrugStock(drugStock),
			FromSupplier: drugStock.Drug.Supplier,
			Procurement:  calculateRecommendation(drugStock, unitsByCode[drugStock.Drug.VmedisCode]),
		}
	}

	data := proxy.DrugProcurementRecommendationsResponse{
		Recommendations: recommendations,
		ComputedAt:      time.Now(),
	}

	log.Printf("Writing %d procurement recommendations to cache\n", len(recommendations))
	dataJson, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshalling procurement recommendations: %s\n", err)
		return
	}

	compressed, err := zlibCompress(dataJson)
	if err != nil {
		log.Printf("Error compressing procurement recommendations: %s\n", err)
		return
	}

	if err := redisClient.Set(ctx, procurementRecommendationsKey, compressed, 7*24*time.Hour).Err(); err != nil {
		log.Printf("Error writing procurement recommendations to cache: %s\n", err)
		return
	}

	log.Printf("Wrote %d procurement recommendations to cache (size: %d bytes) \n", len(recommendations), len(dataJson))
}

// getDrugUnits returns drug units by drug code.
// The drug units are sorted from the smallest to the largest.
func getDrugUnits(db *gorm.DB, drugCodes []string) (map[string][]models.DrugUnit, error) {
	var units []models.DrugUnit
	if err := db.Where("drug_vmedis_code IN ?", drugCodes).Find(&units).Error; err != nil {
		return nil, fmt.Errorf("get drug units from DB: %w", err)
	}

	unitsByCode := make(map[string][]models.DrugUnit, len(units))
	for _, unit := range units {
		unitsByCode[unit.DrugVmedisCode] = append(unitsByCode[unit.DrugVmedisCode], unit)
	}

	for code, units := range unitsByCode {
		sorted := make([]models.DrugUnit, 0, len(units))

		last := ""
		for {
			found := false
			for _, u := range units {
				if u.ParentUnit == last {
					sorted = append(sorted, u)
					last = u.Unit
					found = true
					break
				}
			}

			if !found {
				break
			}
		}

		unitsByCode[code] = sorted
	}

	return unitsByCode, nil
}

func calculateRecommendation(stock client.DrugStock, drugUnits []models.DrugUnit) proxy.Stock {
	smallestQ := stock.Drug.MinimumStock.Quantity*2 - stock.Stock.Quantity
	if smallestQ < 2 {
		smallestQ = 2
	}

	fallback := proxy.Stock{
		Unit:     stock.Stock.Unit,
		Quantity: smallestQ,
	}

	if len(drugUnits) == 0 {
		return fallback
	}

	qPerUnit := make([]float64, len(drugUnits))
	qPerUnit[0] = smallestQ

	for i := 1; i < len(drugUnits); i++ {
		qPerUnit[i] = math.Round(qPerUnit[i-1] / math.Max(drugUnits[i].ConversionToParentUnit, 1))
	}

	for i := len(drugUnits) - 1; i >= 0; i-- {
		if qPerUnit[i] > 0 {
			return proxy.Stock{
				Unit:     drugUnits[i].Unit,
				Quantity: qPerUnit[i],
			}
		}
	}

	return fallback
}

func zlibCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
