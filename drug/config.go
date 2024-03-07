package drug

import (
	"time"
)

type ApiHandlerConfig struct {
	Service                    *Service
	StockOpnameLookupStartDate time.Time
}
