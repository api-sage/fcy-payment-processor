package services_test

import (
	"context"
	"testing"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/services"
)

func TestTransferServiceTransferFundsValidationError(t *testing.T) {
	svc := services.NewTransferService(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		"100100",
		"0123456789",
		"0123456790",
		"0123456791",
		"0123456792",
		"0123456793",
		"0123456794",
		"0123456795",
	)

	_, err := svc.TransferFunds(context.Background(), models.InternalTransferRequest{})
	if err == nil {
		t.Fatal("expected validation error for empty transfer request")
	}
}

