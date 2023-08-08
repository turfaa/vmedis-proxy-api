package proxy

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/vmedis/client"
	"github.com/turfaa/vmedis-proxy-api/vmedis/dumper"
)

const (
	procurementRecommendationsKey = "static_key.procurement_recommendations.json.zlib"
)

// DrugProcurementRecommendationsResponse represents the drug procurement recommendations API response.
type DrugProcurementRecommendationsResponse struct {
	Recommendations []DrugProcurementRecommendation `json:"recommendations"`
	ComputedAt      time.Time                       `json:"computedAt"`
}

// DrugProcurementRecommendation represents one drug procurement recommendation.
type DrugProcurementRecommendation struct {
	DrugStock    `json:",inline"`
	FromSupplier string  `json:"fromSupplier"`
	Procurement  Stock   `json:"procurement"`
	Alternatives []Stock `json:"alternatives"`
}

// DrugStock is the stock of a drug.
type DrugStock struct {
	Drug  Drug  `json:"drug"`
	Stock Stock `json:"stock"`
}

// FromClientDrugStock converts DrugStock from client schema to proxy schema.
func FromClientDrugStock(cd client.DrugStock) DrugStock {
	return DrugStock{
		Drug:  FromClientDrug(cd.Drug),
		Stock: FromClientStock(cd.Stock),
	}
}

// HandleDumpProcurementRecommendations handles the request to calculate and dump the procurement recommendations.
func (s *ApiServer) HandleDumpProcurementRecommendations(c *gin.Context) {
	go dumper.DumpProcurementRecommendations(context.Background(), s.DB, s.RedisClient, s.Client)
	c.JSON(200, gin.H{
		"message": "dumping procurement recommendations",
	})
}

// HandleProcurementRecommendations handles the request to get the procurement recommendations.
func (s *ApiServer) HandleProcurementRecommendations(c *gin.Context) {
	compressed, err := s.RedisClient.Get(c, procurementRecommendationsKey).Result()
	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to get procurement recommendations from Redis: " + err.Error(),
		})
		return
	}

	data, err := zlibDecompress([]byte(compressed))
	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to decompress procurement recommendations: " + err.Error(),
		})
		return
	}

	var response DrugProcurementRecommendationsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		c.JSON(500, gin.H{
			"error": "failed to unmarshal procurement recommendations: " + err.Error(),
		})
		return
	}

	c.JSON(200, response)
}

func zlibDecompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	if _, err := buf.ReadFrom(r); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
