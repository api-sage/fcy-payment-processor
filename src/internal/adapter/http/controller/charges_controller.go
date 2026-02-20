package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/logger"
)

type ChargesService interface {
	GetChargesSummary(ctx context.Context, req models.GetChargesRequest) (models.Response[models.GetChargesResponse], error)
}

type ChargesController struct {
	service ChargesService
}

func NewChargesController(service ChargesService) *ChargesController {
	return &ChargesController{service: service}
}

func (c *ChargesController) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	handler := http.HandlerFunc(c.getCharges)
	if authMiddleware != nil {
		handler = authMiddleware(handler).ServeHTTP
	}

	mux.Handle("/get-charges", http.HandlerFunc(handler))
}

func (c *ChargesController) getCharges(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodGet {
		response := models.ErrorResponse[models.GetChargesResponse]("method not allowed")
		writeJSON(w, http.StatusMethodNotAllowed, response)
		logResponse(r, http.StatusMethodNotAllowed, response, start)
		return
	}

	req := models.GetChargesRequest{
		Amount:       r.URL.Query().Get("amount"),
		FromCurrency: r.URL.Query().Get("fromCurrency"),
	}
	logRequest(r, req)

	response, err := c.service.GetChargesSummary(r.Context(), req)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := http.StatusInternalServerError
		if response.Message == "validation failed" {
			status = http.StatusBadRequest
		}
		if response.Message == "Rate not found for currency pair" {
			status = http.StatusNotFound
		}
		writeJSON(w, status, response)
		logResponse(r, status, response, start)
		return
	}

	writeJSON(w, http.StatusOK, response)
	logResponse(r, http.StatusOK, response, start)
}
