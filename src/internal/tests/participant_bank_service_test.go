package services_test

import (
	"context"
	"testing"

	"github.com/api-sage/fcy-payment-processor/src/internal/domain"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/services"
)

type participantBankRepoStub struct {
	banks []domain.ParticipantBank
	err   error
}

func (s participantBankRepoStub) GetAll(context.Context) ([]domain.ParticipantBank, error) {
	return s.banks, s.err
}

func TestParticipantBankServiceGetParticipantBanksSuccess(t *testing.T) {
	svc := services.NewParticipantBankService(participantBankRepoStub{
		banks: []domain.ParticipantBank{
			{BankName: "Grey", BankCode: "100100"},
			{BankName: "Other", BankCode: "123456"},
		},
	})

	resp, err := svc.GetParticipantBanks(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !resp.Success || resp.Data == nil {
		t.Fatal("expected successful response with data")
	}
	if len(*resp.Data) != 2 {
		t.Fatalf("expected 2 participant banks, got %d", len(*resp.Data))
	}
}

