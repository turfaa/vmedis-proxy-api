package report

import (
	"context"
	"time"

	"github.com/jordan-wright/email"

	"github.com/turfaa/vmedis-proxy-api/procurement"
	"github.com/turfaa/vmedis-proxy-api/sale"
)

type AggregatedProcurementsGetter interface {
	GetAggregatedProcurementsBetweenTime(ctx context.Context, from time.Time, to time.Time) ([]procurement.AggregatedProcurement, error)
}

type AggregatedSalesGetter interface {
	GetAggregatedSalesBetweenTime(ctx context.Context, from time.Time, to time.Time) ([]sale.AggregatedSale, error)
}

type EmailSender interface {
	Send(mail *email.Email, timeout time.Duration) error
}
