package controller

import (
	"context"
	"net/http"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
)

type UserService interface {
	CreateUser(ctx context.Context, req models.CreateUserRequest) (models.Response[models.CreateUserResponse], error)
	GetUser(ctx context.Context, id string) (models.Response[models.GetUserResponse], error)
}

type UserController struct {
	service UserService
}

func NewUserController(service UserService) *UserController {
	return &UserController{service: service}
}

func (c *UserController) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	})
}
