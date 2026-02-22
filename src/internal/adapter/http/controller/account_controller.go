package controller

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/service_interfaces"
)

const (
	createAccountPath = "/create-account"
	getAccountPath    = "/get-account"
	depositFundsPath  = "/deposit-funds"
)

type AccountController struct {
	service service_interfaces.AccountService
}

func NewAccountController(service service_interfaces.AccountService) *AccountController {
	return &AccountController{service: service}
}

func (c *AccountController) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	var createAccountHandler http.Handler = http.HandlerFunc(c.createAccount)
	var getAccountHandler http.Handler = http.HandlerFunc(c.getAccount)
	var depositFundsHandler http.Handler = http.HandlerFunc(c.depositFunds)

	if authMiddleware != nil {
		createAccountHandler = authMiddleware(createAccountHandler)
		getAccountHandler = authMiddleware(getAccountHandler)
		depositFundsHandler = authMiddleware(depositFundsHandler)
	}

	mux.Handle(createAccountPath, createAccountHandler)
	mux.Handle(getAccountPath, getAccountHandler)
	mux.Handle(depositFundsPath, depositFundsHandler)
}

func (c *AccountController) createAccount(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		response := commons.ErrorResponse[models.CreateAccountResponse]("method not allowed")
		c.respondError(w, http.StatusMethodNotAllowed, response, r, start)
		return
	}

	var req models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.CreateAccountResponse]("invalid request body", err.Error())
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	if err := req.Validate(); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.CreateAccountResponse]("validation failed", err.Error())
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	logRequest(r, req)
	response, err := c.service.CreateAccount(r.Context(), req)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := mapResponseToStatus(response.Message)
		c.respondError(w, status, response, r, start)
		return
	}

	c.respondSuccess(w, http.StatusCreated, response, r, start)
}

func (c *AccountController) getAccount(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodGet {
		response := commons.ErrorResponse[models.GetAccountResponse]("method not allowed")
		c.respondError(w, http.StatusMethodNotAllowed, response, r, start)
		return
	}

	accountNumber := strings.TrimSpace(r.URL.Query().Get("accountNumber"))
	bankCode := strings.TrimSpace(r.URL.Query().Get("bankCode"))

	if accountNumber == "" || bankCode == "" {
		response := commons.ErrorResponse[models.GetAccountResponse]("validation failed", "accountNumber and bankCode are required")
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	logRequest(r, map[string]string{
		"accountNumber": accountNumber,
		"bankCode":      bankCode,
	})

	response, err := c.service.GetAccount(r.Context(), accountNumber, bankCode)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := mapResponseToStatus(response.Message)
		c.respondError(w, status, response, r, start)
		return
	}

	c.respondSuccess(w, http.StatusOK, response, r, start)
}

func (c *AccountController) depositFunds(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		response := commons.ErrorResponse[models.DepositFundsResponse]("method not allowed")
		c.respondError(w, http.StatusMethodNotAllowed, response, r, start)
		return
	}

	var req models.DepositFundsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.DepositFundsResponse]("invalid request body", err.Error())
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	if err := req.Validate(); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.DepositFundsResponse]("validation failed", err.Error())
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	logRequest(r, req)
	response, err := c.service.DepositFunds(r.Context(), req)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := mapResponseToStatus(response.Message)
		c.respondError(w, status, response, r, start)
		return
	}

	c.respondSuccess(w, http.StatusOK, response, r, start)
}

// mapResponseToStatus maps response messages to appropriate HTTP status codes
func mapResponseToStatus(message string) int {
	switch message {
	case "validation failed":
		return http.StatusBadRequest
	case "Account not found":
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// respondSuccess sends a successful JSON response with logging
func (c *AccountController) respondSuccess(w http.ResponseWriter, status int, payload any, r *http.Request, start time.Time) {
	if err := writeJSON(w, status, payload); err != nil {
		logError(r, err, logger.Fields{"action": "write response"})
	}
	logResponse(r, status, payload, start)
}

// respondError sends an error JSON response with logging
func (c *AccountController) respondError(w http.ResponseWriter, status int, payload any, r *http.Request, start time.Time) {
	if err := writeJSON(w, status, payload); err != nil {
		logError(r, err, logger.Fields{"action": "write error response"})
	}
	logResponse(r, status, payload, start)
}

// writeJSON encodes and writes JSON response, returning any encoding error
func writeJSON(w http.ResponseWriter, status int, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(payload)
}
