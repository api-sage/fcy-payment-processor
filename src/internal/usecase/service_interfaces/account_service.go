package service_interfaces

import (
	"context"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/commons"
)

type AccountService interface {
	CreateAccount(ctx context.Context, req models.CreateAccountRequest) (commons.Response[models.CreateAccountResponse], error)
	GetAccount(ctx context.Context, accountNumber string, bankCode string) (commons.Response[models.GetAccountResponse], error)
	DepositFunds(ctx context.Context, req models.DepositFundsRequest) (commons.Response[models.DepositFundsResponse], error)
}
