package services_test

import (
	"context"
	"testing"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/services"
	"github.com/shopspring/decimal"
)

func TestAccountServiceCreateAccountValidationError(t *testing.T) {
	svc := services.NewAccountService(nil, nil, nil, "100100")

	_, err := svc.CreateAccount(context.Background(), models.CreateAccountRequest{})
	if err == nil {
		t.Fatal("expected validation error for empty create account request")
	}
}

func TestAccountServiceGetAccountValidationError(t *testing.T) {
	svc := services.NewAccountService(nil, nil, nil, "100100")

	_, err := svc.GetAccount(context.Background(), "", "100100")
	if err == nil {
		t.Fatal("expected validation error for missing account number")
	}
}

func TestAccountServiceDepositFundsValidationError(t *testing.T) {
	svc := services.NewAccountService(nil, nil, nil, "100100")

	_, err := svc.DepositFunds(context.Background(), models.DepositFundsRequest{
		AccountNumber: "123",
		Amount:        decimal.Zero,
	})
	if err == nil {
		t.Fatal("expected validation error for invalid deposit request")
	}
}

