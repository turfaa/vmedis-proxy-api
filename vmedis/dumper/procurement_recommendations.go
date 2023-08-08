package dumper

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/proxy"
)

const (
	procurementRecommendationsKey = "static_key.procurement_recommendations.json.zlib"
)

// DumpProcurementRecommendations calculates and dumps procurement recommendations to cache.
func DumpProcurementRecommendations(redisClient *redis.Client, vmedisClient *client.Client) {
	log.Println("Computing procurement recommendations and writing them to cache....")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
	defer cancel()

	log.Printf("Getting all out of stock drugs for writing procurement recommendations")
	oosDrugs, err := vmedisClient.GetAllOutOfStockDrugs(ctx)
	if err != nil {
		log.Printf("Error getting all out of stock drugs for writing procurement recommendations: %s\n", err)
		return
	}

	log.Printf("Got %d out of stock drugs for writing procurement recommendations\n", len(oosDrugs))

	recommendations := make([]proxy.DrugProcurementRecommendation, len(oosDrugs))
	for i, drugStock := range oosDrugs {
		q := drugStock.Drug.MinimumStock.Quantity*2 - drugStock.Stock.Quantity
		if q < 2 {
			q = 2
		}

		recommendations[i] = proxy.DrugProcurementRecommendation{
			DrugStock:    proxy.FromClientDrugStock(drugStock),
			FromSupplier: drugStock.Drug.Supplier,
			Procurement: proxy.Stock{
				Unit:     drugStock.Stock.Unit,
				Quantity: q,
			},
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
