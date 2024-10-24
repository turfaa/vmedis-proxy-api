package shift

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/turfaa/vmedis-proxy-api/money"

	"github.com/gin-gonic/gin"
	"github.com/turfaa/vmedis-proxy-api/cui"
	"github.com/turfaa/vmedis-proxy-api/pkg2/slices2"
	"github.com/turfaa/vmedis-proxy-api/pkg2/time2"
)

type ApiHandler struct {
	service *Service
}

func NewApiHandler(service *Service) *ApiHandler {
	return &ApiHandler{service: service}
}

func (h *ApiHandler) GetShiftByVmedisID(c *gin.Context) {
	vmedisID, err := strconv.Atoi(c.Param("vmedis_id"))
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("invalid vmedis id: %s", err)})
		return
	}

	shift, err := h.service.GetShiftByVmedisID(c, vmedisID)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to get shift by vmedis id %d: %s", vmedisID, err)})
		return
	}

	c.JSON(200, h.transformShiftToTable(shift))
}

func (h *ApiHandler) transformShiftToTable(shift Shift) cui.Table {
	return cui.Table{
		Rows: []cui.Row{
			{
				ID: "kode",
				Columns: []string{
					"Kode",
					shift.Code,
				},
			},
			{
				ID: "kasir",
				Columns: []string{
					"Kasir",
					shift.Cashier,
				},
			},
			{
				ID: "mulai",
				Columns: []string{
					"Mulai",
					time2.FormatDateTime(shift.StartedAt),
				},
			},
			{
				ID: "selesai",
				Columns: []string{
					"Selesai",
					time2.FormatDateTime(shift.EndedAt),
				},
			},
			{
				ID: "kas_awal",
				Columns: []string{
					"Kas Awal",
					money.FormatRupiah(shift.InitialCash),
				},
			},
			{
				ID: "kas_akhir_seharusnya",
				Columns: []string{
					"Kas Akhir Seharusnya",
					money.FormatRupiah(shift.ExpectedFinalCash),
				},
			},
			{
				ID: "kas_akhir_sebenarnya",
				Columns: []string{
					"Kas Akhir Sebenarnya",
					money.FormatRupiah(shift.ActualFinalCash),
				},
			},
			{
				ID: "selisih",
				Columns: []string{
					"Selisih",
					money.FormatRupiah(shift.FinalCashDifference),
				},
			},
			{
				ID: "catatan",
				Columns: []string{
					"Catatan",
					shift.Notes,
				},
			},
		},
	}
}

func (h *ApiHandler) GetShifts(c *gin.Context) {
	from, to, err := time2.GetTimeRangeFromQuery(c)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("invalid time range: %s", err)})
		return
	}

	shifts, err := h.service.GetShiftsBetween(c, from, to)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to get shifts: %s", err)})
		return
	}

	c.JSON(200, h.transformShiftsToTable(shifts))
}

func (h *ApiHandler) DumpShiftsFromVmedisToDB(c *gin.Context) {
	var req DumpShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if !errors.Is(err, io.EOF) {
			c.JSON(400, gin.H{"error": fmt.Sprintf("invalid request: %s", err)})
			return
		}
	}

	from, to, err := time2.ParseTimeRange("", req.From, req.To)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("invalid time range: %s", err)})
		return
	}

	shifts, err := h.service.DumpShiftsFromVmedisToDB(c, from, to)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to dump shifts from vmedis to db: %s", err)})
		return
	}

	c.JSON(200, h.transformShiftsToTable(shifts))
}

func (h *ApiHandler) transformShiftsToTable(shifts []Shift) cui.Table {
	header := []string{
		"Kode",
		"Kasir",
		"Mulai",
		"Selesai",
		"Kas Awal",
		"Kas Akhir Seharusnya",
		"Kas Akhir Sebenarnya",
		"Selisih",
		"Catatan",
	}

	rows := slices2.Map(shifts, func(shift Shift) cui.Row {
		return cui.Row{
			ID: strconv.Itoa(shift.VmedisID),
			Columns: []string{
				shift.Code,
				shift.Cashier,
				time2.FormatDateTime(shift.StartedAt),
				time2.FormatDateTime(shift.EndedAt),
				money.FormatRupiah(shift.InitialCash),
				money.FormatRupiah(shift.ExpectedFinalCash),
				money.FormatRupiah(shift.ActualFinalCash),
				money.FormatRupiah(shift.FinalCashDifference),
				shift.Notes,
			},
		}
	})

	return cui.Table{
		Header: header,
		Rows:   rows,
	}
}
