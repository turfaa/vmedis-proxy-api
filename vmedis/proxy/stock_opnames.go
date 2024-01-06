package proxy

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"

	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis/dumper"
	"github.com/turfaa/vmedis-proxy-api/vmedis/proxy/schema"
)

// HandleGetStockOpnames handles the request to get the stock opnames.
func (s *ApiServer) HandleGetStockOpnames(c *gin.Context) {
	sos, err := s.getStockOpnames(c)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get stock opnames: %s", err),
		})
		return
	}

	c.JSON(200, schema.StockOpnamesResponse{StockOpnames: sos})
}

// HandleGetStockOpnameSummaries handles the request to get the stock opname summary.
func (s *ApiServer) HandleGetStockOpnameSummaries(c *gin.Context) {
	sos, err := s.getStockOpnames(c)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get stock opnames: %s", err),
		})
		return
	}

	summaries := make([]schema.StockOpnameSummary, 0, len(sos))

	var currentDrug []schema.StockOpname
	for _, so := range sos {
		if len(currentDrug) == 0 || stockOpnameShouldSummarizeTogether(so, currentDrug[0]) {
			currentDrug = append(currentDrug, so)
			continue
		}

		summaries = append(summaries, summarizeOneDrugStockOpnames(currentDrug))
		currentDrug = append(currentDrug[:0], so)
	}

	if len(currentDrug) > 0 {
		summaries = append(summaries, summarizeOneDrugStockOpnames(currentDrug))
	}

	c.JSON(200, schema.StockOpnameSummariesResponse{Summaries: summaries})
}

// HandleDumpStockOpnames handles the request to dump the stock opnames.
func (s *ApiServer) HandleDumpStockOpnames(c *gin.Context) {
	go dumper.DumpDailyStockOpnames(context.Background(), s.DB, s.Client)
	c.JSON(200, gin.H{
		"message": "dumping stock opnames",
	})
}

func (s *ApiServer) getStockOpnames(c *gin.Context) ([]schema.StockOpname, error) {
	timeFrom, timeUntil, err := getTimeRangeFromQuery(c)
	if err != nil {
		return nil, fmt.Errorf("get dates from query: %w", err)
	}

	var stockOpnames []models.StockOpname
	if err := s.DB.Where("date >= ? AND date <= ?", datatypes.Date(timeFrom), datatypes.Date(timeUntil)).Order("vmedis_id").Find(&stockOpnames).Error; err != nil {
		return nil, fmt.Errorf("get stock opnames from database: %w", err)
	}

	sos := make([]schema.StockOpname, len(stockOpnames))
	for i, so := range stockOpnames {
		sos[i] = schema.FromModelsStockOpname(so)
	}

	return sos, nil
}

// summarizeOneDrugStockOpnames assumes that the stock opnames are:
// - In the same date
// - Sorted chronologically
// - For the same drug
func summarizeOneDrugStockOpnames(stockOpnames []schema.StockOpname) schema.StockOpnameSummary {
	summary := schema.StockOpnameSummary{
		Date:     stockOpnames[0].Date,
		DrugCode: stockOpnames[0].DrugCode,
		DrugName: stockOpnames[0].DrugName,
		Unit:     stockOpnames[0].Unit,
		Changes:  make([]schema.StockChange, 0, len(stockOpnames)),
	}

	for _, so := range stockOpnames {
		if so.InitialQuantity == so.RealQuantity {
			continue
		}

		summary.QuantityDifference += so.QuantityDifference
		summary.HPPDifference += so.HPPDifference
		summary.SalePriceDifference += so.SalePriceDifference

		summary.Changes = append(summary.Changes, schema.StockChange{
			BatchCode:       so.BatchCode,
			InitialQuantity: so.InitialQuantity,
			RealQuantity:    so.RealQuantity,
		})
	}

	return summary
}

func stockOpnameShouldSummarizeTogether(so1, so2 schema.StockOpname) bool {
	return so1.Date == so2.Date && so1.DrugCode == so2.DrugCode && so1.Unit == so2.Unit
}
