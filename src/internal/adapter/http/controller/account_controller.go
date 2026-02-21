package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/commons"
	"github.com/api-sage/ccy-payment-processor/src/internal/logger"
)

type AccountService interface {
	CreateAccount(ctx context.Context, req models.CreateAccountRequest) (commons.Response[models.CreateAccountResponse], error)
	GetAccount(ctx context.Context, accountNumber string, bankCode string) (commons.Response[models.GetAccountResponse], error)
	DepositFunds(ctx context.Context, req models.DepositFundsRequest) (commons.Response[models.DepositFundsResponse], error)
}

type AccountController struct {
	service AccountService
}

func NewAccountController(service AccountService) *AccountController {
	return &AccountController{service: service}
}

func (c *AccountController) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	handler := http.HandlerFunc(c.createAccount)
	getAccountHandler := http.HandlerFunc(c.getAccount)
	depositFundsHandler := http.HandlerFunc(c.depositFunds)
	if authMiddleware != nil {
		handler = authMiddleware(handler).ServeHTTP
		getAccountHandler = authMiddleware(getAccountHandler).ServeHTTP
		depositFundsHandler = authMiddleware(depositFundsHandler).ServeHTTP
	}
	mux.Handle("/create-account", http.HandlerFunc(handler))
	mux.Handle("/get-account", http.HandlerFunc(getAccountHandler))
	mux.Handle("/deposit-funds", http.HandlerFunc(depositFundsHandler))
}

func (c *AccountController) createAccount(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logRequest(r, nil)

	if r.Method != http.MethodPost {
		response := commons.ErrorResponse[models.CreateAccountResponse]("method not allowed")
		writeJSON(w, http.StatusMethodNotAllowed, response)
		logResponse(r, http.StatusMethodNotAllowed, response, start)
		return
	}

	var req models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.CreateAccountResponse]("invalid request body", err.Error())
		writeJSON(w, http.StatusBadRequest, response)
		logResponse(r, http.StatusBadRequest, response, start)
		return
	}
	logRequest(r, req)

	if err := req.Validate(); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.CreateAccountResponse]("validation failed", err.Error())
		writeJSON(w, http.StatusBadRequest, response)
		logResponse(r, http.StatusBadRequest, response, start)
		return
	}

	response, err := c.service.CreateAccount(r.Context(), req)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := http.StatusInternalServerError
		if response.Message == "validation failed" {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, response)
		logResponse(r, status, response, start)
		return
	}

	writeJSON(w, http.StatusCreated, response)
	logResponse(r, http.StatusCreated, response, start)
}

func (c *AccountController) getAccount(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	if r.Method != http.MethodGet {
		response := commons.ErrorResponse[models.GetAccountResponse]("method not allowed")
		writeJSON(w, http.StatusMethodNotAllowed, response)
		logResponse(r, http.StatusMethodNotAllowed, response, start)
		return
	}

	accountNumber := r.URL.Query().Get("accountNumber")
	bankCode := r.URL.Query().Get("bankCode")
	logRequest(r, map[string]string{
		"accountNumber": accountNumber,
		"bankCode":      bankCode,
	})
	response, err := c.service.GetAccount(r.Context(), accountNumber, bankCode)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := http.StatusInternalServerError
		if response.Message == "validation failed" {
			status = http.StatusBadRequest
		}
		if response.Message == "Account not found" {
			status = http.StatusNotFound
		}
		writeJSON(w, status, response)
		logResponse(r, status, response, start)
		return
	}

	writeJSON(w, http.StatusOK, response)
	logResponse(r, http.StatusOK, response, start)
}

func (c *AccountController) depositFunds(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logRequest(r, nil)

	if r.Method != http.MethodPost {
		response := commons.ErrorResponse[models.DepositFundsResponse]("method not allowed")
		writeJSON(w, http.StatusMethodNotAllowed, response)
		logResponse(r, http.StatusMethodNotAllowed, response, start)
		return
	}

	var req models.DepositFundsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.DepositFundsResponse]("invalid request body", err.Error())
		writeJSON(w, http.StatusBadRequest, response)
		logResponse(r, http.StatusBadRequest, response, start)
		return
	}
	logRequest(r, req)

	response, err := c.service.DepositFunds(r.Context(), req)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := http.StatusInternalServerError
		if response.Message == "validation failed" {
			status = http.StatusBadRequest
		}
		if response.Message == "Account not found" {
			status = http.StatusNotFound
		}
		writeJSON(w, status, response)
		logResponse(r, status, response, start)
		return
	}

	writeJSON(w, http.StatusOK, response)
	logResponse(r, http.StatusOK, response, start)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
