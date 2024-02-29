package report

import (
	"context"
	"time"

	"github.com/jordan-wright/email"

	"github.com/turfaa/vmedis-proxy-api/procurement"
)

type AggregatedProcurementsGetter interface {
	GetAggregatedProcurementsBetweenTime(ctx context.Context, from time.Time, to time.Time) ([]procurement.AggregatedProcurement, error)
}

type EmailSender interface {
	Send(mail *email.Email, timeout time.Duration) error
}
