package controller

import (
	"net/http"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/service_interfaces"
)

const getTransactionPath = "/get-transaction"

type TransactionController struct {
	service service_interfaces.TransactionService
}

func NewTransactionController(service service_interfaces.TransactionService) *TransactionController {
	return &TransactionController{service: service}
}

func (c *TransactionController) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	var handler http.Handler = http.HandlerFunc(c.getTransaction)

	if authMiddleware != nil {
		handler = authMiddleware(handler)
	}

	mux.Handle(getTransactionPath, handler)
}

func (c *TransactionController) getTransaction(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodGet {
		response := commons.ErrorResponse[models.GetTransactionResponse]("method not allowed")
		c.respondError(w, http.StatusMethodNotAllowed, response, r, start)
		return
	}

	req := models.GetTransactionRequest{
		TransactionReference: r.URL.Query().Get("transactionReference"),
	}

	logRequest(r, req)
	response, err := c.service.GetTransaction(r.Context(), req)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := mapTransactionResponseToStatus(response.Message)
		c.respondError(w, status, response, r, start)
		return
	}

	c.respondSuccess(w, http.StatusOK, response, r, start)
}

// mapTransactionResponseToStatus maps service response messages to HTTP status codes
func mapTransactionResponseToStatus(message string) int {
	switch message {
	case "validation failed":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// respondSuccess sends a successful JSON response with logging
func (c *TransactionController) respondSuccess(w http.ResponseWriter, status int, payload any, r *http.Request, start time.Time) {
	if err := writeJSON(w, status, payload); err != nil {
		logError(r, err, logger.Fields{"action": "write response"})
	}
	logResponse(r, status, payload, start)
}

// respondError sends an error JSON response with logging
func (c *TransactionController) respondError(w http.ResponseWriter, status int, payload any, r *http.Request, start time.Time) {
	if err := writeJSON(w, status, payload); err != nil {
		logError(r, err, logger.Fields{"action": "write error response"})
	}
	logResponse(r, status, payload, start)
}
