package vmedis

import (
	"net/url"
	"strconv"
	"time"
)

type SearchByTimeParameters[T ParameterType] struct {
	StartTime time.Time
	EndTime   time.Time
	Page      int
}

func (p SearchByTimeParameters[T]) ToQuery(timeFormat string) string {
	values := make(url.Values, 4)
	values.Add("page", strconv.Itoa(p.Page))

	if !p.StartTime.IsZero() || !p.EndTime.IsZero() {
		var t T
		values.Add(t.QueryLabel()+"[cari]", "4")
		values.Add(t.QueryLabel()+"[tanggalawal]", p.StartTime.Format(timeFormat))
		values.Add(t.QueryLabel()+"[tanggalakhir]", p.EndTime.Format(timeFormat))
	}

	return values.Encode()
}

type ParameterType interface {
	QueryLabel() string
}

type ParameterTypeProcurements struct{}

func (ParameterTypeProcurements) QueryLabel() string {
	return "LapPembelianObatBatchSearch"
}

type ParameterTypeShifts struct{}

func (ParameterTypeShifts) QueryLabel() string {
	return "LaporangantishiftSearch"
}
