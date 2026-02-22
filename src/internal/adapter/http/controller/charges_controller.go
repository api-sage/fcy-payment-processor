package controller

import (
	"net/http"
	"strings"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/service_interfaces"
	"github.com/shopspring/decimal"
)

const (
	getChargesPath = "/get-charges"
)

type ChargesController struct {
	service service_interfaces.ChargesService
}

func NewChargesController(service service_interfaces.ChargesService) *ChargesController {
	return &ChargesController{service: service}
}

func (c *ChargesController) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	var handler http.Handler = http.HandlerFunc(c.getCharges)
	if authMiddleware != nil {
		handler = authMiddleware(handler)
	}

	mux.Handle(getChargesPath, handler)
}

func (c *ChargesController) getCharges(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodGet {
		response := commons.ErrorResponse[models.GetChargesResponse]("method not allowed")
		c.respondError(w, http.StatusMethodNotAllowed, response, r, start)
		return
	}

	amountRaw := strings.TrimSpace(r.URL.Query().Get("amount"))
	if amountRaw == "" {
		response := commons.ErrorResponse[models.GetChargesResponse]("validation failed", "amount is required")
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	amount, err := decimal.NewFromString(amountRaw)
	if err != nil {
		logError(r, err, logger.Fields{"field": "amount"})
		response := commons.ErrorResponse[models.GetChargesResponse]("validation failed", "amount must be a valid decimal number")
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	fromCurrency := strings.TrimSpace(r.URL.Query().Get("fromCurrency"))
	if fromCurrency == "" {
		response := commons.ErrorResponse[models.GetChargesResponse]("validation failed", "fromCurrency is required")
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	req := models.GetChargesRequest{
		Amount:       amount,
		FromCurrency: fromCurrency,
	}
	logRequest(r, req)

	response, err := c.service.GetChargesSummary(r.Context(), req)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := mapChargesResponseToStatus(response.Message)
		c.respondError(w, status, response, r, start)
		return
	}

	c.respondSuccess(w, http.StatusOK, response, r, start)
}

// mapChargesResponseToStatus maps charges response messages to appropriate HTTP status codes
func mapChargesResponseToStatus(message string) int {
	switch message {
	case "validation failed":
		return http.StatusBadRequest
	case "Rate not found for currency pair":
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// respondSuccess sends a successful JSON response with logging
func (c *ChargesController) respondSuccess(w http.ResponseWriter, status int, payload any, r *http.Request, start time.Time) {
	if err := writeJSON(w, status, payload); err != nil {
		logError(r, err, logger.Fields{"action": "write response"})
	}
	logResponse(r, status, payload, start)
}

// respondError sends an error JSON response with logging
func (c *ChargesController) respondError(w http.ResponseWriter, status int, payload any, r *http.Request, start time.Time) {
	if err := writeJSON(w, status, payload); err != nil {
		logError(r, err, logger.Fields{"action": "write error response"})
	}
	logResponse(r, status, payload, start)
}
