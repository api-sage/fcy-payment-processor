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
	getRatesPath         = "/get-rates"
	getRatePath          = "/get-rate"
	convertFCYAmountPath = "/convert-fcy-amount"
)

type RateController struct {
	service service_interfaces.RateService
}

func NewRateController(service service_interfaces.RateService) *RateController {
	return &RateController{service: service}
}

func (c *RateController) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	var getRatesHandler http.Handler = http.HandlerFunc(c.getRates)
	var getRateHandler http.Handler = http.HandlerFunc(c.getRate)
	var getCcyRatesHandler http.Handler = http.HandlerFunc(c.getCcyRates)

	if authMiddleware != nil {
		getRatesHandler = authMiddleware(getRatesHandler)
		getRateHandler = authMiddleware(getRateHandler)
		getCcyRatesHandler = authMiddleware(getCcyRatesHandler)
	}

	mux.Handle(getRatesPath, getRatesHandler)
	mux.Handle(getRatePath, getRateHandler)
	mux.Handle(convertFCYAmountPath, getCcyRatesHandler)
}

func (c *RateController) getRates(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodGet {
		response := commons.ErrorResponse[[]models.RateResponse]("method not allowed")
		c.respondError(w, http.StatusMethodNotAllowed, response, r, start)
		return
	}

	logRequest(r, nil)
	response, err := c.service.GetRates(r.Context())
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		c.respondError(w, http.StatusInternalServerError, response, r, start)
		return
	}

	c.respondSuccess(w, http.StatusOK, response, r, start)
}

func (c *RateController) getRate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		response := commons.ErrorResponse[models.RateResponse]("method not allowed")
		c.respondError(w, http.StatusMethodNotAllowed, response, r, start)
		return
	}

	var req models.GetRateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.RateResponse]("invalid request body", err.Error())
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	if err := req.Validate(); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.RateResponse]("validation failed", err.Error())
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	logRequest(r, req)
	response, err := c.service.GetRate(r.Context(), req)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := mapRateResponseToStatus(response.Message)
		c.respondError(w, status, response, r, start)
		return
	}

	c.respondSuccess(w, http.StatusOK, response, r, start)
}

func (c *RateController) getCcyRates(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		response := commons.ErrorResponse[models.GetCcyRatesResponse]("method not allowed")
		c.respondError(w, http.StatusMethodNotAllowed, response, r, start)
		return
	}

	var req models.GetCcyRatesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.GetCcyRatesResponse]("invalid request body", err.Error())
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	if err := req.Validate(); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.GetCcyRatesResponse]("validation failed", err.Error())
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	logRequest(r, req)
	response, err := c.service.GetCcyRates(r.Context(), req)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := mapRateResponseToStatus(response.Message)
		c.respondError(w, status, response, r, start)
		return
	}

	c.respondSuccess(w, http.StatusOK, response, r, start)
}

// mapRateResponseToStatus maps rate response messages to appropriate HTTP status codes
func mapRateResponseToStatus(message string) int {
	switch message {
	case "validation failed":
		return http.StatusBadRequest
	case "Rate not found", "Rate not found for currency pair":
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// respondSuccess sends a successful JSON response with logging
func (c *RateController) respondSuccess(w http.ResponseWriter, status int, payload any, r *http.Request, start time.Time) {
	if err := writeJSON(w, status, payload); err != nil {
		logError(r, err, logger.Fields{"action": "write response"})
	}
	logResponse(r, status, payload, start)
}

// respondError sends an error JSON response with logging
func (c *RateController) respondError(w http.ResponseWriter, status int, payload any, r *http.Request, start time.Time) {
	if err := writeJSON(w, status, payload); err != nil {
		logError(r, err, logger.Fields{"action": "write error response"})
	}
	logResponse(r, status, payload, start)
}
