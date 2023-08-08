package proxy

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/go-redis/redis/v8"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
)

const (
	procurementRecommendationsInterval = time.Hour
	procurementRecommendationsKey      = "static_key.procurement_recommendations.json"
)

func runCacheWriter(redisClient *redis.Client, vmedisClient *client.Client) func() {
	scheduler := gocron.NewScheduler(time.Local)

	if _, err := scheduler.Every(procurementRecommendationsInterval).Do(writeProcurementRecommendations, redisClient, vmedisClient); err != nil {
		log.Fatalf("Error scheduling procurement recommendations writer: %s\n", err)
	}

	scheduler.StartAsync()
	return scheduler.Stop
}

func writeProcurementRecommendations(redisClient *redis.Client, vmedisClient *client.Client) {
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

	recommendations := make([]DrugProcurementRecommendation, len(oosDrugs))
	for i, drugStock := range oosDrugs {
		recommendations[i] = DrugProcurementRecommendation{
			DrugStock:    FromClientDrugStock(drugStock),
			FromSupplier: drugStock.Drug.Supplier,
			Procurement: Stock{
				Unit:     drugStock.Stock.Unit,
				Quantity: drugStock.Drug.MinimumStock.Quantity*2 - drugStock.Stock.Quantity,
			},
		}
	}

	data := DrugProcurementRecommendationsResponse{
		Recommendations: recommendations,
		ComputedAt:      time.Now(),
	}

	log.Printf("Writing %d procurement recommendations to cache\n", len(recommendations))
	dataJson, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshalling procurement recommendations: %s\n", err)
		return
	}

	if err := redisClient.Set(ctx, procurementRecommendationsKey, dataJson, 0).Err(); err != nil {
		log.Printf("Error writing procurement recommendations to cache: %s\n", err)
		return
	}

	log.Printf("Wrote %d procurement recommendations to cache (size: %d bytes) \n", len(recommendations), len(dataJson))
}
