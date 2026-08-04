package sale

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/turfaa/vmedis-proxy-api/cui"
	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/money"
)

const defaultLastDrugSalesLimit = 5

func (s *ApiHandler) GetLastDrugSales(c *gin.Context) {
	var request LastDrugSalesRequest
	if err := c.ShouldBindUri(&request); err != nil {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("invalid request: %s", err),
		})
		return
	}

	if err := c.ShouldBindQuery(&request); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("invalid request: %s", err),
		})
		return
	}

	limit := request.Limit
	if limit <= 0 {
		limit = defaultLastDrugSalesLimit
	}

	sales, err := s.service.GetLastDrugSales(c.Request.Context(), request.DrugCode, limit)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get last drug sales: %s", err),
		})
		return
	}

	c.JSON(200, s.transformLastDrugSalesToTable(sales))
}

func (s *ApiHandler) transformLastDrugSalesToTable(sales []DrugSale) cui.Table {
	header := []string{
		"Tanggal Penjualan",
		"Nomor Faktur",
		"Jumlah",
		"Harga Satuan",
		"Total",
		"Kategori Harga",
	}

	rows := make([]cui.Row, len(sales))
	for i, sale := range sales {
		rows[i] = cui.Row{
			ID: strconv.Itoa(i),
			Columns: []string{
				sale.SoldAt.Format("2006-01-02 15:04"),
				sale.InvoiceNumber,
				drug.Stock{
					Quantity: sale.Amount,
					Unit:     sale.Unit,
				}.String(),
				money.FormatRupiah(sale.UnitPrice) + " / " + sale.Unit,
				money.FormatRupiah(sale.Total),
				sale.PriceCategory,
			},
		}
	}

	return cui.Table{
		Header: header,
		Rows:   rows,
	}
}
