package proxy

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/dumper"
	"github.com/turfaa/vmedis-proxy-api/proxy/schema"
)

const (
	procurementRecommendationsKey = "static_key.procurement_recommendations.json.zlib"
)

// HandleDumpProcurementRecommendations handles the request to calculate and dump the procurement recommendations.
func (s *ApiServer) HandleDumpProcurementRecommendations(c *gin.Context) {
	go dumper.DumpProcurementRecommendations(context.Background(), s.db, s.redisClient, s.client)
	c.JSON(200, gin.H{
		"message": "dumping procurement recommendations",
	})
}

// HandleProcurementRecommendations handles the request to get the procurement recommendations.
func (s *ApiServer) HandleProcurementRecommendations(c *gin.Context) {
	compressed, err := s.redisClient.Get(c, procurementRecommendationsKey).Result()
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

	var response schema.DrugProcurementRecommendationsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		c.JSON(500, gin.H{
			"error": "failed to unmarshal procurement recommendations: " + err.Error(),
		})
		return
	}

	c.JSON(200, response)
}