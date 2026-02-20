package services

import (
	"context"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
	"github.com/api-sage/ccy-payment-processor/src/internal/logger"
)

type ParticipantBankService struct {
	participantBankRepo domain.ParticipantBankRepository
}

func NewParticipantBankService(participantBankRepo domain.ParticipantBankRepository) *ParticipantBankService {
	return &ParticipantBankService{participantBankRepo: participantBankRepo}
}

func (s *ParticipantBankService) GetParticipantBanks(ctx context.Context) (models.Response[[]models.ParticipantBankResponse], error) {
	logger.Info("participant bank service get participant banks request", nil)

	banks, err := s.participantBankRepo.GetAll(ctx)
	if err != nil {
		logger.Error("participant bank service get participant banks failed", err, nil)
		return models.ErrorResponse[[]models.ParticipantBankResponse]("failed to fetch participant banks", "Unable to fetch participant banks right now"), err
	}

	resp := make([]models.ParticipantBankResponse, 0, len(banks))
	for _, bank := range banks {
		resp = append(resp, models.ParticipantBankResponse{
			BankName: bank.BankName,
			BankCode: bank.BankCode,
		})
	}

	logger.Info("participant bank service get participant banks success", logger.Fields{
		"count": len(resp),
	})

	return models.SuccessResponse("participant banks fetched successfully", resp), nil
}
