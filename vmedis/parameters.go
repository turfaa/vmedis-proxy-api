package vmedis

import (
	"net/url"
	"strconv"
	"time"
)

type GetProcurementsParameters struct {
	StartDate time.Time
	EndDate   time.Time
	Page      int
}

func (p GetProcurementsParameters) ToQuery() string {
	values := make(url.Values, 4)
	values.Add("page", strconv.Itoa(p.Page))

	if !p.StartDate.IsZero() || !p.EndDate.IsZero() {
		values.Add("LapPembelianObatBatchSearch[cari]", "4")
		values.Add("LapPembelianObatBatchSearch[tanggalawal]", p.StartDate.Format(dateFormat))
		values.Add("LapPembelianObatBatchSearch[tanggalakhir]", p.EndDate.Format(dateFormat))
	}

	return values.Encode()
}
