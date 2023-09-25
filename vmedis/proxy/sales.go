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

// HandleGetSales handles the GET /sales endpoint.
func (s *ApiServer) HandleGetSales(c *gin.Context) {
	salesModels, err := s.getSales(c)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get sales: %s", err),
		})
		return
	}

	sales := make([]schema.Sale, len(salesModels))
	for i, sale := range salesModels {
		sales[i] = schema.FromModelsSale(sale)
	}

	c.JSON(200, schema.SalesResponse{Sales: sales})
}

// HandleGetSoldDrugs handles the GET /sold-drugs endpoint.
func (s *ApiServer) HandleGetSoldDrugs(c *gin.Context) {
	salesModels, err := s.getSales(c)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get sales: %s", err),
		})
		return
	}

	var (
		drugCodes []string
		drugs     = make(map[string]schema.SoldDrug)
	)
	for _, sale := range salesModels {
		for _, saleUnit := range sale.SaleUnits {
			drugCodes = append(drugCodes, saleUnit.DrugCode)
			drugs[saleUnit.DrugCode] = schema.SoldDrug{
				Occurrences: drugs[saleUnit.DrugCode].Occurrences + 1,
				TotalAmount: drugs[saleUnit.DrugCode].TotalAmount + saleUnit.Total,
			}
		}
	}

	var drugsModels []models.Drug
	if err := s.DB.Find(&drugsModels, "vmedis_code IN ?", drugCodes).Error; err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get drugs: %s", err),
		})
		return
	}

	for _, drug := range drugsModels {
		drugs[drug.VmedisCode] = schema.SoldDrug{
			Drug:        schema.FromModelsDrug(drug),
			Occurrences: drugs[drug.VmedisCode].Occurrences,
			TotalAmount: drugs[drug.VmedisCode].TotalAmount,
		}
	}

	drugsSlice := make([]schema.SoldDrug, 0, len(drugs))
	for _, drug := range drugs {
		drugsSlice = append(drugsSlice, drug)
	}

	date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	c.JSON(200, schema.SoldDrugsResponse{Drugs: drugsSlice, Date: date})
}

// HandleDumpSales handles the request to dump today's sales.
func (s *ApiServer) HandleDumpSales(c *gin.Context) {
	go dumper.DumpDailySales(context.Background(), s.DB, s.Client)

	c.JSON(200, gin.H{
		"message": "dumping today's sales",
	})
}

func (s *ApiServer) getSales(c *gin.Context) ([]models.Sale, error) {
	date := c.Query("date")

	var from, until time.Time
	if date == "" {
		from, until = today()
	} else {
		var err error
		from, until, err = day(date)
		if err != nil {
			return nil, fmt.Errorf("get day from date query [%s]: %w", date, err)
		}
	}

	var salesModels []models.Sale
	if err := s.DB.Preload("SaleUnits").Find(&salesModels, "sold_at >= ? AND sold_at <= ?", from, until).Error; err != nil {
		return nil, fmt.Errorf("get sales from database: %w", err)
	}

	return salesModels, nil
}
