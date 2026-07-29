package procurement

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/turfaa/vmedis-proxy-api/cui"
	"github.com/turfaa/vmedis-proxy-api/drug"
	"github.com/turfaa/vmedis-proxy-api/money"
	"github.com/turfaa/vmedis-proxy-api/pkg2/time2"

	"github.com/gin-gonic/gin"
)

func (h *ApiHandler) GetSupplierProcurementRecaps(c *gin.Context) {
	from, until, err := time2.GetTimeRangeFromQuery(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("invalid request: %s", err),
		})
		return
	}

	recaps, err := h.service.GetSupplierProcurementRecapsBetweenTime(c.Request.Context(), from, until)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get supplier procurement recaps: %s", err),
		})
		return
	}

	c.JSON(200, h.transformSupplierProcurementRecapsToTable(recaps))
}

func (h *ApiHandler) transformSupplierProcurementRecapsToTable(recaps []SupplierProcurementRecap) cui.Table {
	header := []string{
		"No",
		"Supplier",
		"Jumlah Faktur",
		"Total Pembelian",
	}

	var (
		totalInvoiceCount int64
		totalAmount       float64
	)

	rows := make([]cui.Row, len(recaps))
	for i, recap := range recaps {
		totalInvoiceCount += recap.InvoiceCount
		totalAmount += recap.Total

		rows[i] = cui.Row{
			ID: strconv.Itoa(i),
			Columns: []string{
				strconv.Itoa(i + 1),
				recap.Supplier,
				formatInvoiceCount(recap.InvoiceCount),
				money.FormatRupiah(recap.Total),
			},
		}
	}

	return cui.Table{
		Header: header,
		Rows:   rows,
		Footer: []string{
			"",
			"Total",
			formatInvoiceCount(totalInvoiceCount),
			money.FormatRupiah(totalAmount),
		},
	}
}

func formatInvoiceCount(count int64) string {
	return strconv.FormatInt(count, 10) + " Faktur"
}

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

	procurements, err := h.service.GetLastDrugProcurements(c.Request.Context(), request.DrugCode, limit)
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
