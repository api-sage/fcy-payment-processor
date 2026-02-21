package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
)

type RateService interface {
	GetRates(ctx context.Context) (commons.Response[[]models.RateResponse], error)
	GetRate(ctx context.Context, req models.GetRateRequest) (commons.Response[models.RateResponse], error)
	GetCcyRates(ctx context.Context, req models.GetCcyRatesRequest) (commons.Response[models.GetCcyRatesResponse], error)
}

type RateController struct {
	service RateService
}

func NewRateController(service RateService) *RateController {
	return &RateController{service: service}
}

func (c *RateController) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	getRatesHandler := http.HandlerFunc(c.getRates)
	getRateHandler := http.HandlerFunc(c.getRate)
	getCcyRatesHandler := http.HandlerFunc(c.getCcyRates)

	if authMiddleware != nil {
		getRatesHandler = authMiddleware(getRatesHandler).ServeHTTP
		getRateHandler = authMiddleware(getRateHandler).ServeHTTP
		getCcyRatesHandler = authMiddleware(getCcyRatesHandler).ServeHTTP
	}

	mux.Handle("/get-rates", http.HandlerFunc(getRatesHandler))
	mux.Handle("/get-rate", http.HandlerFunc(getRateHandler))
	mux.Handle("/convert-fcy-amount", http.HandlerFunc(getCcyRatesHandler))
}

func (c *RateController) getRates(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logRequest(r, nil)

	if r.Method != http.MethodGet {
		response := commons.ErrorResponse[[]models.RateResponse]("method not allowed")
		writeJSON(w, http.StatusMethodNotAllowed, response)
		logResponse(r, http.StatusMethodNotAllowed, response, start)
		return
	}

	response, err := c.service.GetRates(r.Context())
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		writeJSON(w, http.StatusInternalServerError, response)
		logResponse(r, http.StatusInternalServerError, response, start)
		return
	}

	writeJSON(w, http.StatusOK, response)
	logResponse(r, http.StatusOK, response, start)
}

func (c *RateController) getRate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logRequest(r, nil)

	if r.Method != http.MethodPost {
		response := commons.ErrorResponse[models.RateResponse]("method not allowed")
		writeJSON(w, http.StatusMethodNotAllowed, response)
		logResponse(r, http.StatusMethodNotAllowed, response, start)
		return
	}

	var req models.GetRateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.RateResponse]("invalid request body", err.Error())
		writeJSON(w, http.StatusBadRequest, response)
		logResponse(r, http.StatusBadRequest, response, start)
		return
	}
	logRequest(r, req)

	response, err := c.service.GetRate(r.Context(), req)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := http.StatusInternalServerError
		if response.Message == "validation failed" {
			status = http.StatusBadRequest
		}
		if response.Message == "Rate not found" {
			status = http.StatusNotFound
		}
		writeJSON(w, status, response)
		logResponse(r, status, response, start)
		return
	}

	writeJSON(w, http.StatusOK, response)
	logResponse(r, http.StatusOK, response, start)
}

func (c *RateController) getCcyRates(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logRequest(r, nil)

	if r.Method != http.MethodPost {
		response := commons.ErrorResponse[models.GetCcyRatesResponse]("method not allowed")
		writeJSON(w, http.StatusMethodNotAllowed, response)
		logResponse(r, http.StatusMethodNotAllowed, response, start)
		return
	}

	var req models.GetCcyRatesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.GetCcyRatesResponse]("invalid request body", err.Error())
		writeJSON(w, http.StatusBadRequest, response)
		logResponse(r, http.StatusBadRequest, response, start)
		return
	}
	logRequest(r, req)

	response, err := c.service.GetCcyRates(r.Context(), req)
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
