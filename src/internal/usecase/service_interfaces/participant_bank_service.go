package service_interfaces

import (
	"context"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/commons"
)

type ParticipantBankService interface {
	GetParticipantBanks(ctx context.Context) (commons.Response[[]models.ParticipantBankResponse], error)
}
