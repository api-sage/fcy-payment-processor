package domain

import "time"

type Rate struct {
	ID           int64
	FromCurrency string
	ToCurrency   string
	SellRate     string
	BuyRate      string
	RateDate     time.Time
	CreatedAt    time.Time
}
