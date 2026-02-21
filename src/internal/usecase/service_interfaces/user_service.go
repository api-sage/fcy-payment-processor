package service_interfaces

import (
	"context"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
)

type UserService interface {
	CreateUser(ctx context.Context, req models.CreateUserRequest) (commons.Response[models.CreateUserResponse], error)
	GetUser(ctx context.Context, id string) (commons.Response[models.GetUserResponse], error)
	VerifyUserPin(ctx context.Context, customerID string, pin string) (commons.Response[models.VerifyUserPinResponse], error)
}
