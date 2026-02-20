package domain

import "time"

type Rate struct {
	ID           int64
	FromCurrency string
	ToCurrency   string
	Rate         string
	RateDate     time.Time
	CreatedAt    time.Time
}
