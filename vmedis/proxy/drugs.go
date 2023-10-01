package proxy

import (
	"context"
	"fmt"
	"time"

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

// HandleGetDrugsToStockOpname handles the request to get the drugs to stock opname.
func (s *ApiServer) HandleGetDrugsToStockOpname(c *gin.Context) {
	todayFrom, todayUntil, err := getDatesFromQuery(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("failed to parse date query: %s", err),
		})
		return
	}

	yesterdayFrom, yesterdayUntil := todayFrom.Add(-24*time.Hour), todayUntil.Add(-24*time.Hour)

	yesterdaySales, err := s.getSalesBetween(yesterdayFrom, yesterdayUntil)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get sales between %s and %s: %s", yesterdayFrom, yesterdayUntil, err),
		})
		return
	}

	lastMonthFrom := todayFrom.AddDate(0, -1, 0)

	var alreadyStockOpnamedDrugCodes []string
	if err := s.DB.Model(&models.StockOpname{}).Where("date BETWEEN ? AND ?", lastMonthFrom, todayUntil).Pluck("drug_code", &alreadyStockOpnamedDrugCodes).Error; err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get already stock opnamed drug codes: %s", err),
		})
		return
	}

	drugsToStockOpnameCodesMap := make(map[string]struct{})
	for _, sale := range yesterdaySales {
		for _, saleUnit := range sale.SaleUnits {
			drugsToStockOpnameCodesMap[saleUnit.DrugCode] = struct{}{}
		}
	}

	for _, drugCode := range alreadyStockOpnamedDrugCodes {
		delete(drugsToStockOpnameCodesMap, drugCode)
	}

	drugsToStockOpnameCodes := make([]string, 0, len(drugsToStockOpnameCodesMap))
	for drugCode := range drugsToStockOpnameCodesMap {
		drugsToStockOpnameCodes = append(drugsToStockOpnameCodes, drugCode)
	}

	var drugsToStockOpnameModels []models.Drug
	if err := s.DB.Preload("Units").Find(&drugsToStockOpnameModels, "vmedis_code IN ?", drugsToStockOpnameCodes).Error; err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get drugs to stock opname: %s", err),
		})
		return
	}

	drugsToStockOpname := make([]schema.Drug, len(drugsToStockOpnameModels))
	for i, drug := range drugsToStockOpnameModels {
		drugsToStockOpname[i] = schema.FromModelsDrug(drug)
	}

	c.JSON(200, schema.DrugsResponse{Drugs: drugsToStockOpname})
}

// HandleDumpDrugs handles the request to dump the drugs.
func (s *ApiServer) HandleDumpDrugs(c *gin.Context) {
	go dumper.DumpDrugs(context.Background(), s.DB, s.Client, s.DrugDetailsPuller)
	c.JSON(200, gin.H{
		"message": "dumping drugs",
	})
}
