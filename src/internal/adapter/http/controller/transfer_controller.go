package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/logger"
)

type TransferService interface {
	TransferFunds(ctx context.Context, req models.InternalTransferRequest) (models.Response[models.InternalTransferResponse], error)
}

type TransferController struct {
	service TransferService
}

func NewTransferController(service TransferService) *TransferController {
	return &TransferController{service: service}
}

func (c *TransferController) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	handler := http.HandlerFunc(c.transfer)
	if authMiddleware != nil {
		handler = authMiddleware(handler).ServeHTTP
	}

	mux.Handle("/transfer-funds", http.HandlerFunc(handler))
}

func (c *TransferController) transfer(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logRequest(r, nil)

	if r.Method != http.MethodPost {
		response := models.ErrorResponse[models.InternalTransferResponse]("method not allowed")
		writeJSON(w, http.StatusMethodNotAllowed, response)
		logResponse(r, http.StatusMethodNotAllowed, response, start)
		return
	}

	var req models.InternalTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError(r, err, nil)
		response := models.ErrorResponse[models.InternalTransferResponse]("invalid request body", err.Error())
		writeJSON(w, http.StatusBadRequest, response)
		logResponse(r, http.StatusBadRequest, response, start)
		return
	}
	logRequest(r, req)

	response, err := c.service.TransferFunds(r.Context(), req)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := http.StatusInternalServerError
		if response.Message == "validation failed" {
			status = http.StatusBadRequest
		}
		if response.Message == "Debit account not found" || response.Message == "Credit account not found" || response.Message == "Rate not found" {
			status = http.StatusNotFound
		}
		if response.Message == "Insufficient balance" {
			status = http.StatusUnprocessableEntity
		}

		writeJSON(w, status, response)
		logResponse(r, status, response, start)
		return
	}

	writeJSON(w, http.StatusOK, response)
	logResponse(r, http.StatusOK, response, start)
}
