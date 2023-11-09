package proxy

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/turfaa/vmedis-proxy-api/vmedis/database/models"
	"github.com/turfaa/vmedis-proxy-api/vmedis/dumper"
	"github.com/turfaa/vmedis-proxy-api/vmedis/proxy/schema"
)

var (
	drugsUpdatedAtThresholds = []time.Duration{
		24*time.Hour + 3*time.Hour,
		3 * 24 * time.Hour,
		7 * 24 * time.Hour,
		30 * 24 * time.Hour,
		9999 * 24 * time.Hour,
	}
)

// HandleGetDrugs handles the request to get the drugs.
func (s *ApiServer) HandleGetDrugs(c *gin.Context) {
	drugs, err := s.getDrugs(nil, 1000)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get drugs: %s", err),
		})
		return
	}

	for i, drug := range drugs {
		drug.Units = filterUnits(drug.Units)
		drugs[i] = drug
	}

	c.JSON(200, schema.DrugsResponse{Drugs: drugs})
}

// HandleGetDrugsToStockOpname handles the request to get the drugs to stock opname.
func (s *ApiServer) HandleGetDrugsToStockOpname(c *gin.Context) {
	mode := strings.ToLower(c.DefaultQuery("mode", "sales-based"))

	switch mode {
	case "sales-based":
		s.HandleGetSalesBasedDrugsToStockOpname(c)

	case "conservative":
		s.HandleGetConservativeDrugsToStockOpname(c)

	default:
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("invalid mode: %s", mode),
		})
	}
}

// HandleGetSalesBasedDrugsToStockOpname handles the request to get the drugs to stock opname based on sales in the last month.
func (s *ApiServer) HandleGetSalesBasedDrugsToStockOpname(c *gin.Context) {
	todayFrom, todayUntil, err := getOneDayFromQuery(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("failed to parse date query: %s", err),
		})
		return
	}

	lastMonthFrom, yesterdayUntil := todayFrom.AddDate(0, -1, 0), todayUntil.Add(-24*time.Hour)

	type salesStat struct {
		DrugCode    string
		Occurrences int
		TotalAmount float64
	}

	var salesStats []salesStat
	if err := s.DB.Raw("SELECT drug_code, COUNT(*) AS occurrences, SUM(total) AS total_amount FROM sale_units WHERE invoice_number IN (SELECT invoice_number FROM sales WHERE sold_at BETWEEN ? AND ?) GROUP BY drug_code", lastMonthFrom, yesterdayUntil).Find(&salesStats).Error; err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get sales stats: %s", err),
		})
		return
	}

	var alreadyStockOpnamedDrugCodes []string
	if err := s.DB.Model(&models.StockOpname{}).Where("date BETWEEN ? AND ?", lastMonthFrom, todayUntil).Pluck("drug_code", &alreadyStockOpnamedDrugCodes).Error; err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get already stock opnamed drug codes: %s", err),
		})
		return
	}

	drugsToStockOpnameCodesMap := make(map[string]struct{})
	drugSales := make(map[string]salesStat)
	for _, stats := range salesStats {
		drugsToStockOpnameCodesMap[stats.DrugCode] = struct{}{}
		drugSales[stats.DrugCode] = stats
	}

	for _, drugCode := range alreadyStockOpnamedDrugCodes {
		delete(drugsToStockOpnameCodesMap, drugCode)
	}

	drugsToStockOpnameCodes := make([]string, 0, len(drugsToStockOpnameCodesMap))
	for drugCode := range drugsToStockOpnameCodesMap {
		drugsToStockOpnameCodes = append(drugsToStockOpnameCodes, drugCode)
	}

	drugsToStockOpname, err := s.getDrugs(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("vmedis_code IN ?", drugsToStockOpnameCodes)
	}, 0)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get drugs to stock opname: %s", err),
		})
		return
	}

	sort.Slice(drugsToStockOpname, func(i, j int) bool {
		if drugSales[drugsToStockOpname[i].VmedisCode].Occurrences == drugSales[drugsToStockOpname[j].VmedisCode].Occurrences {
			return drugSales[drugsToStockOpname[i].VmedisCode].TotalAmount > drugSales[drugsToStockOpname[j].VmedisCode].TotalAmount
		}

		return drugSales[drugsToStockOpname[i].VmedisCode].Occurrences > drugSales[drugsToStockOpname[j].VmedisCode].Occurrences
	})

	c.JSON(200, schema.DrugsResponse{Drugs: drugsToStockOpname})
}

// HandleGetConservativeDrugsToStockOpname handles the request to get the drugs to stock opname based on all drugs.
func (s *ApiServer) HandleGetConservativeDrugsToStockOpname(c *gin.Context) {
	todayFrom, todayUntil, err := getOneDayFromQuery(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("failed to parse date query: %s", err),
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

	drugsToStockOpname, err := s.getDrugs(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("vmedis_code NOT IN ?", alreadyStockOpnamedDrugCodes)
	}, 0)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get drugs to stock opname: %s", err),
		})
		return
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

func (s *ApiServer) getDrugs(additionalQuery func(tx *gorm.DB) *gorm.DB, drugsThreshold int) ([]schema.Drug, error) {
	if additionalQuery == nil {
		additionalQuery = func(tx *gorm.DB) *gorm.DB { return tx }
	}

	var drugs []models.Drug
	for _, updatedAtThreshold := range drugsUpdatedAtThresholds {
		if err := additionalQuery(s.DB.Preload("Units").Where("updated_at >= ?", time.Now().Add(-updatedAtThreshold)).Order("name")).Find(&drugs).Error; err != nil {
			return nil, fmt.Errorf("get drugs: %w", err)
		}

		if len(drugs) > drugsThreshold {
			break
		}
	}

	d := make([]schema.Drug, len(drugs))
	for i, drug := range drugs {
		d[i] = schema.FromModelsDrug(drug)
	}

	return d, nil
}
