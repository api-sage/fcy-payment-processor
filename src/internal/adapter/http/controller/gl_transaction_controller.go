package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/service_interfaces"
)

const getGLTransactionsPath = "/get-gl-transactions"

type GLTransactionController struct {
	service service_interfaces.GLTransactionService
}

func NewGLTransactionController(service service_interfaces.GLTransactionService) *GLTransactionController {
	return &GLTransactionController{service: service}
}

func (c *GLTransactionController) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	var handler http.Handler = http.HandlerFunc(c.getGLTransactions)

	if authMiddleware != nil {
		handler = authMiddleware(handler)
	}

	mux.Handle(getGLTransactionsPath, handler)
}

func (c *GLTransactionController) getGLTransactions(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodGet {
		response := commons.ErrorResponse[models.GetGLTransactionsResponse]("method not allowed")
		c.respondError(w, http.StatusMethodNotAllowed, response, r, start)
		return
	}

	req, err := parseGLTransactionsQuery(r)
	if err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.GetGLTransactionsResponse]("validation failed", err.Error())
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	logRequest(r, req)
	response, err := c.service.GetGLTransactions(r.Context(), req)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := mapGLTransactionResponseToStatus(response.Message)
		c.respondError(w, status, response, r, start)
		return
	}

	c.respondSuccess(w, http.StatusOK, response, r, start)
}

// parseGLTransactionsQuery reads the query string. page defaults to 1 and pageSize to the default page size.
// Values that are not integers are rejected here. Range checks happen in the request's Validate.
func parseGLTransactionsQuery(r *http.Request) (models.GetGLTransactionsRequest, error) {
	query := r.URL.Query()

	req := models.GetGLTransactionsRequest{
		StartDate:       query.Get("startDate"),
		EndDate:         query.Get("endDate"),
		GLAccountNumber: query.Get("glAccountNumber"),
		Page:            1,
		PageSize:        models.GLTransactionsDefaultPageSize,
	}

	if raw := query.Get("page"); raw != "" {
		page, err := strconv.Atoi(raw)
		if err != nil {
			return req, fmt.Errorf("page must be an integer")
		}
		req.Page = page
	}

	if raw := query.Get("pageSize"); raw != "" {
		pageSize, err := strconv.Atoi(raw)
		if err != nil {
			return req, fmt.Errorf("pageSize must be an integer")
		}
		req.PageSize = pageSize
	}

	return req, nil
}

// mapGLTransactionResponseToStatus maps service response messages to HTTP status codes
func mapGLTransactionResponseToStatus(message string) int {
	switch message {
	case "validation failed":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// respondSuccess sends a successful JSON response with logging
func (c *GLTransactionController) respondSuccess(w http.ResponseWriter, status int, payload any, r *http.Request, start time.Time) {
	if err := writeJSON(w, status, payload); err != nil {
		logError(r, err, logger.Fields{"action": "write response"})
	}
	logResponse(r, status, payload, start)
}

// respondError sends an error JSON response with logging
func (c *GLTransactionController) respondError(w http.ResponseWriter, status int, payload any, r *http.Request, start time.Time) {
	if err := writeJSON(w, status, payload); err != nil {
		logError(r, err, logger.Fields{"action": "write error response"})
	}
	logResponse(r, status, payload, start)
}
