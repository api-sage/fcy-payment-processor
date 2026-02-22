package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/service_interfaces"
)

const (
	transferFundsPath = "/transfer-funds"
)

type TransferController struct {
	service service_interfaces.TransferService
}

func NewTransferController(service service_interfaces.TransferService) *TransferController {
	return &TransferController{service: service}
}

func (c *TransferController) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	var handler http.Handler = http.HandlerFunc(c.transfer)
	if authMiddleware != nil {
		handler = authMiddleware(handler)
	}

	mux.Handle(transferFundsPath, handler)
}

func (c *TransferController) transfer(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		response := commons.ErrorResponse[models.InternalTransferResponse]("method not allowed")
		c.respondError(w, http.StatusMethodNotAllowed, response, r, start)
		return
	}

	var req models.InternalTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.InternalTransferResponse]("invalid request body", err.Error())
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	if err := req.Validate(); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error())
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	logRequest(r, req)
	response, err := c.service.TransferFunds(r.Context(), req)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := mapTransferResponseToStatus(response.Message)
		c.respondError(w, status, response, r, start)
		return
	}

	c.respondSuccess(w, http.StatusOK, response, r, start)
}

// mapTransferResponseToStatus maps transfer response messages to appropriate HTTP status codes
func mapTransferResponseToStatus(message string) int {
	switch message {
	case "validation failed":
		return http.StatusBadRequest
	case "Debit account not found", "Credit account not found", "Rate not found":
		return http.StatusNotFound
	case "Insufficient balance":
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// respondSuccess sends a successful JSON response with logging
func (c *TransferController) respondSuccess(w http.ResponseWriter, status int, payload any, r *http.Request, start time.Time) {
	if err := writeJSON(w, status, payload); err != nil {
		logError(r, err, logger.Fields{"action": "write response"})
	}
	logResponse(r, status, payload, start)
}

// respondError sends an error JSON response with logging
func (c *TransferController) respondError(w http.ResponseWriter, status int, payload any, r *http.Request, start time.Time) {
	if err := writeJSON(w, status, payload); err != nil {
		logError(r, err, logger.Fields{"action": "write error response"})
	}
	logResponse(r, status, payload, start)
}
