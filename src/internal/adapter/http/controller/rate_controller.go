package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
)

type RateService interface {
	GetRates(ctx context.Context) (models.Response[[]models.RateResponse], error)
	GetRate(ctx context.Context, req models.GetRateRequest) (models.Response[models.RateResponse], error)
	GetCcyRates(ctx context.Context, req models.GetCcyRatesRequest) (models.Response[models.GetCcyRatesResponse], error)
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
	mux.Handle("/convertfcyamount", http.HandlerFunc(getCcyRatesHandler))
}

func (c *RateController) getRates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse[[]models.RateResponse]("method not allowed"))
		return
	}

	response, err := c.service.GetRates(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (c *RateController) getRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse[models.RateResponse]("method not allowed"))
		return
	}

	var req models.GetRateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse[models.RateResponse]("invalid request body", err.Error()))
		return
	}

	response, err := c.service.GetRate(r.Context(), req)
	if err != nil {
		status := http.StatusInternalServerError
		if response.Message == "validation failed" {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, response)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (c *RateController) getCcyRates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse[models.GetCcyRatesResponse]("method not allowed"))
		return
	}

	var req models.GetCcyRatesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse[models.GetCcyRatesResponse]("invalid request body", err.Error()))
		return
	}

	response, err := c.service.GetCcyRates(r.Context(), req)
	if err != nil {
		status := http.StatusInternalServerError
		if response.Message == "validation failed" {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, response)
		return
	}

	writeJSON(w, http.StatusOK, response)
}
