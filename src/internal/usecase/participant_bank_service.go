package usecase

import (
	"context"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
)

type ParticipantBankService struct {
	participantBankRepo domain.ParticipantBankRepository
}

func NewParticipantBankService(participantBankRepo domain.ParticipantBankRepository) *ParticipantBankService {
	return &ParticipantBankService{participantBankRepo: participantBankRepo}
}

func (s *ParticipantBankService) GetParticipantBanks(ctx context.Context) (models.Response[[]models.ParticipantBankResponse], error) {
	banks, err := s.participantBankRepo.GetAll(ctx)
	if err != nil {
		return models.ErrorResponse[[]models.ParticipantBankResponse]("failed to fetch participant banks", err.Error()), err
	}

	resp := make([]models.ParticipantBankResponse, 0, len(banks))
	for _, bank := range banks {
		resp = append(resp, models.ParticipantBankResponse{
			BankName: bank.BankName,
			BankCode: bank.BankCode,
		})
	}

	return models.SuccessResponse("participant banks fetched successfully", resp), nil
}
