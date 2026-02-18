package controller

import (
	"context"
	"encoding/json"
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
	handler := http.HandlerFunc(c.createUser)
	if authMiddleware != nil {
		handler = authMiddleware(handler).ServeHTTP
	}
	mux.Handle("/users", http.HandlerFunc(handler))
}

func (c *UserController) createUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse[models.CreateUserResponse]("method not allowed"))
		return
	}

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse[models.CreateUserResponse]("invalid request body", err.Error()))
		return
	}

	if err := req.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse[models.CreateUserResponse]("validation failed", err.Error()))
		return
	}

	response, err := c.service.CreateUser(r.Context(), req)
	if err != nil {
		status := http.StatusInternalServerError
		if response.Message == "validation failed" {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, response)
		return
	}

	writeJSON(w, http.StatusCreated, response)
}
