package domain

import "time"

type TransientAccount struct {
	ID                 string
	AccountNumber      string
	AccountDescription string
	Currency           string
	AvailableBalance   string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
