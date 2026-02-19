package memory

import (
	"context"

	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
)

type ParticipantBankRepository struct{}

func NewParticipantBankRepository() *ParticipantBankRepository {
	return &ParticipantBankRepository{}
}

func (r *ParticipantBankRepository) GetAll(_ context.Context) ([]domain.ParticipantBank, error) {
	banks := []domain.ParticipantBank{
		{BankName: "Access Bank", BankCode: "044001"},
		{BankName: "First Bank of Nigeria", BankCode: "011001"},
		{BankName: "Guaranty Trust Bank", BankCode: "058001"},
		{BankName: "United Bank for Africa", BankCode: "033001"},
		{BankName: "Zenith Bank", BankCode: "057001"},
		{BankName: "Fidelity Bank", BankCode: "070001"},
		{BankName: "Ecobank Nigeria", BankCode: "050001"},
		{BankName: "FCMB", BankCode: "214001"},
		{BankName: "Union Bank", BankCode: "032001"},
		{BankName: "Sterling Bank", BankCode: "232001"},
	}

	return banks, nil
}
