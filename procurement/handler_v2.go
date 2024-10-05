package procurement

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/turfaa/vmedis-proxy-api/cui"
	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/money"

	"github.com/gin-gonic/gin"
)

func (h *ApiHandler) GetLastDrugProcurements(c *gin.Context) {
	var request LastDrugProcurementsRequest
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
		limit = 5
	}

	procurements, err := h.service.GetLastDrugProcurements(c, request.DrugCode, limit)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get last drug procurements: %s", err),
		})
		return
	}

	c.JSON(200, h.transformLastDrugProcurementsToTable(procurements))
}

func (h *ApiHandler) transformLastDrugProcurementsToTable(procurements []DrugProcurement) cui.Table {
	header := []string{
		"Tanggal Diinput",
		"Nomor Faktur",
		"Tanggal Faktur",
		"Jumlah",
		"Harga Satuan",
		"Supplier",
	}

	rows := make([]cui.Row, len(procurements))
	for i, procurement := range procurements {
		rows[i] = cui.Row{
			ID: strconv.Itoa(i),
			Columns: []string{
				procurement.CreatedAt.Format("2006-01-02"),
				procurement.InvoiceNumber,
				procurement.InvoiceDate.Format("2006-01-02"),
				drug.Stock{
					Quantity: procurement.Amount,
					Unit:     procurement.Unit,
				}.String(),
				money.FormatRupiah(procurement.TotalUnitPrice) + " / " + procurement.Unit,
				procurement.Supplier,
			},
		}
	}

	return cui.Table{
		Header: header,
		Rows:   rows,
	}
}
