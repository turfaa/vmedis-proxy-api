package proxy

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis/dumper"
	"github.com/turfaa/vmedis-proxy-api/vmedis/proxy/schema"
)

// HandleGetDrugs handles the request to get the drugs.
func (s *ApiServer) HandleGetDrugs(c *gin.Context) {
	var drugs []models.Drug
	if err := s.DB.Preload("Units").Order("name").Find(&drugs).Error; err != nil {
		c.JSON(500, gin.H{
			"error": "failed to get drugs from database: " + err.Error(),
		})
		return
	}

	var res schema.DrugsResponse
	for _, drug := range drugs {
		d := schema.FromModelsDrug(drug)
		d.Units = filterUnits(d.Units)

		res.Drugs = append(res.Drugs, d)
	}

	c.JSON(200, res)
}

// HandleGetDrugsToStockOpname handles the request to get the drugs to stock opname.
func (s *ApiServer) HandleGetDrugsToStockOpname(c *gin.Context) {
	todayFrom, todayUntil, err := getOneDayFromQuery(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("failed to parse date query: %s", err),
		})
		return
	}

	yesterdayFrom, lastMonthFrom := todayFrom.Add(-24*time.Hour), todayFrom.AddDate(0, -1, 0)

	sales, err := s.getSalesBetween(yesterdayFrom, lastMonthFrom)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get sales between %s and %s: %s", yesterdayFrom, lastMonthFrom, err),
		})
		return
	}

	lastThreeMonthsFrom := lastMonthFrom.AddDate(0, -3, 0)

	var alreadyStockOpnamedDrugCodes []string
	if err := s.DB.Model(&models.StockOpname{}).Where("date BETWEEN ? AND ?", lastThreeMonthsFrom, todayUntil).Pluck("drug_code", &alreadyStockOpnamedDrugCodes).Error; err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get already stock opnamed drug codes: %s", err),
		})
		return
	}

	drugsToStockOpnameCodesMap := make(map[string]struct{})
	drugSales := make(map[string]schema.SoldDrug)
	for _, sale := range sales {
		for _, saleUnit := range sale.SaleUnits {
			drugsToStockOpnameCodesMap[saleUnit.DrugCode] = struct{}{}
			drugSales[saleUnit.DrugCode] = schema.SoldDrug{
				Occurrences: drugSales[saleUnit.DrugCode].Occurrences + 1,
				TotalAmount: drugSales[saleUnit.DrugCode].TotalAmount + saleUnit.Total,
			}
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

	sort.Slice(drugsToStockOpname, func(i, j int) bool {
		if drugSales[drugsToStockOpname[i].VmedisCode].Occurrences == drugSales[drugsToStockOpname[j].VmedisCode].Occurrences {
			return drugSales[drugsToStockOpname[i].VmedisCode].TotalAmount > drugSales[drugsToStockOpname[j].VmedisCode].TotalAmount
		}

		return drugSales[drugsToStockOpname[i].VmedisCode].Occurrences > drugSales[drugsToStockOpname[j].VmedisCode].Occurrences
	})

	c.JSON(200, schema.DrugsResponse{Drugs: drugsToStockOpname})
}

// HandleDumpDrugs handles the request to dump the drugs.
func (s *ApiServer) HandleDumpDrugs(c *gin.Context) {
	go dumper.DumpDrugs(context.Background(), s.DB, s.Client, s.DrugDetailsPuller)
	c.JSON(200, gin.H{
		"message": "dumping drugs",
	})
}
