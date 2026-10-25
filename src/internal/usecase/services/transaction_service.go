package services

import (
	"context"
	"strings"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/repository/repo_interfaces"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/service_interfaces"
)

// Verify that TransactionService implements the service_interfaces.TransactionService interface
var _ service_interfaces.TransactionService = (*TransactionService)(nil)

type TransactionService struct {
	transactionRepo repo_interfaces.TransientAccountTransactionRepository
}

func NewTransactionService(transactionRepo repo_interfaces.TransientAccountTransactionRepository) *TransactionService {
	return &TransactionService{transactionRepo: transactionRepo}
}

func (s *TransactionService) GetTransaction(ctx context.Context, req models.GetTransactionRequest) (commons.Response[models.GetTransactionResponse], error) {
	logger.Info("transaction service get transaction request", logger.Fields{
		"payload": logger.SanitizePayload(req),
	})

	if err := req.Validate(); err != nil {
		logger.Error("transaction service get transaction validation failed", err, nil)
		return commons.ErrorResponse[models.GetTransactionResponse]("validation failed", err.Error()), err
	}

	reference := strings.TrimSpace(req.TransactionReference)
	entries, err := s.transactionRepo.FindByExternalReference(ctx, reference)
	if err != nil {
		logger.Error("transaction service get transaction failed", err, nil)
		return commons.ErrorResponse[models.GetTransactionResponse]("failed to get transaction", "Unable to fetch transactions right now"), err
	}

	items := make([]models.GLTransactionResponse, 0, len(entries))
	for _, entry := range entries {
		items = append(items, mapTransientAccountTransactionToResponse(entry))
	}

	response := models.GetTransactionResponse{
		TransactionReference: reference,
		Items:                items,
		Total:                len(items),
	}

	logger.Info("transaction service get transaction success", logger.Fields{
		"count": len(items),
	})

	return commons.SuccessResponse("Transactions fetched successfully", response), nil
}
