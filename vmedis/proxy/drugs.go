package proxy

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis/dumper"
	"github.com/turfaa/vmedis-proxy-api/vmedis/proxy/schema"
)

// HandleGetDrugs handles the request to get the drugs.
func (s *ApiServer) HandleGetDrugs(c *gin.Context) {
	var drugs []models.Drug
	if err := s.DB.Preload("Units").Find(&drugs).Error; err != nil {
		c.JSON(500, gin.H{
			"error": "failed to get drugs from database: " + err.Error(),
		})
		return
	}

	var res schema.DrugsResponse
	for _, drug := range drugs {
		res.Drugs = append(res.Drugs, schema.FromModelsDrug(drug))
	}

	c.JSON(200, res)
}

// HandleDumpDrugs handles the request to dump the drugs.
func (s *ApiServer) HandleDumpDrugs(c *gin.Context) {
	go dumper.DumpDrugs(context.Background(), s.DB, s.Client, s.DrugDetailsPuller)
	c.JSON(200, gin.H{
		"message": "dumping drugs",
	})
}
