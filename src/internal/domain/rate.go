package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Rate struct {
	ID           int64
	FromCurrency string
	ToCurrency   string
	Rate         decimal.Decimal
	RateDate     time.Time
	CreatedAt    time.Time
}
