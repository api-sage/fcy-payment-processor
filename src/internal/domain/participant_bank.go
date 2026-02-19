package domain

import "context"

type ParticipantBank struct {
	BankName string
	BankCode string
}

type ParticipantBankRepository interface {
	GetAll(ctx context.Context) ([]ParticipantBank, error)
}
