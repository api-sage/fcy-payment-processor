package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
)

type AccountService interface {
	CreateAccount(ctx context.Context, req models.CreateAccountRequest) (models.Response[models.CreateAccountResponse], error)
}

type AccountController struct {
	service AccountService
}

func NewAccountController(service AccountService) *AccountController {
	return &AccountController{service: service}
}

func (c *AccountController) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	handler := http.HandlerFunc(c.createAccount)
	if authMiddleware != nil {
		handler = authMiddleware(handler).ServeHTTP
	}
	mux.Handle("/accounts", http.HandlerFunc(handler))
}

func (c *AccountController) createAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse[models.CreateAccountResponse]("method not allowed"))
		return
	}

	var req models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse[models.CreateAccountResponse]("invalid request body", err.Error()))
		return
	}

	if err := req.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse[models.CreateAccountResponse]("validation failed", err.Error()))
		return
	}

	response, err := c.service.CreateAccount(r.Context(), req)
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

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
