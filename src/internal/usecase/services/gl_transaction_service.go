package services

import (
	"context"
	"strings"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/repository/repo_interfaces"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
	"github.com/api-sage/fcy-payment-processor/src/internal/domain"
	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/service_interfaces"
)

// Verify that GLTransactionService implements the service_interfaces.GLTransactionService interface
var _ service_interfaces.GLTransactionService = (*GLTransactionService)(nil)

type GLTransactionService struct {
	transactionRepo repo_interfaces.TransientAccountTransactionRepository
}

func NewGLTransactionService(transactionRepo repo_interfaces.TransientAccountTransactionRepository) *GLTransactionService {
	return &GLTransactionService{transactionRepo: transactionRepo}
}

func (s *GLTransactionService) GetGLTransactions(ctx context.Context, req models.GetGLTransactionsRequest) (commons.Response[models.GetGLTransactionsResponse], error) {
	logger.Info("gl transaction service get gl transactions request", logger.Fields{
		"payload": logger.SanitizePayload(req),
	})

	if err := req.Validate(); err != nil {
		logger.Error("gl transaction service get gl transactions validation failed", err, nil)
		return commons.ErrorResponse[models.GetGLTransactionsResponse]("validation failed", err.Error()), err
	}

	start, end, err := req.DateRange()
	if err != nil {
		return commons.ErrorResponse[models.GetGLTransactionsResponse]("validation failed", err.Error()), err
	}

	entries, total, err := s.transactionRepo.List(ctx, domain.TransientAccountTransactionFilter{
		StartDate:       start,
		EndDate:         end,
		GLAccountNumber: strings.TrimSpace(req.GLAccountNumber),
		Limit:           req.PageSize,
		Offset:          (req.Page - 1) * req.PageSize,
	})
	if err != nil {
		logger.Error("gl transaction service get gl transactions failed", err, nil)
		return commons.ErrorResponse[models.GetGLTransactionsResponse]("failed to get gl transactions", "Unable to fetch GL transactions right now"), err
	}

	items := make([]models.GLTransactionResponse, 0, len(entries))
	for _, entry := range entries {
		items = append(items, mapTransientAccountTransactionToResponse(entry))
	}

	pageSize := int64(req.PageSize)
	response := models.GetGLTransactionsResponse{
		Items:      items,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: int((total + pageSize - 1) / pageSize),
	}

	logger.Info("gl transaction service get gl transactions success", logger.Fields{
		"count": len(items),
		"total": total,
	})

	return commons.SuccessResponse("GL transactions fetched successfully", response), nil
}

func mapTransientAccountTransactionToResponse(entry domain.TransientAccountTransaction) models.GLTransactionResponse {
	return models.GLTransactionResponse{
		ID:                entry.ID,
		TransferID:        entry.TransferID,
		ExternalReference: entry.ExternalRefernece,
		DebitedAccount:    entry.DebitedAccount,
		CreditedAccount:   entry.CreditedAccount,
		EntryType:         string(entry.EntryType),
		Currency:          entry.Currency,
		Amount:            entry.Amount,
		CreatedAt:         entry.CreatedAt.Format(time.RFC3339),
	}
}
