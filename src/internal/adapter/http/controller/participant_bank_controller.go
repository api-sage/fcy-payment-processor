package controller

import (
	"net/http"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/service_interfaces"
)

const (
	getParticipantBanksPath = "/get-participant-banks"
)

type ParticipantBankController struct {
	service service_interfaces.ParticipantBankService
}

func NewParticipantBankController(service service_interfaces.ParticipantBankService) *ParticipantBankController {
	return &ParticipantBankController{service: service}
}

func (c *ParticipantBankController) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	var handler http.Handler = http.HandlerFunc(c.getParticipantBanks)
	if authMiddleware != nil {
		handler = authMiddleware(handler)
	}

	mux.Handle(getParticipantBanksPath, handler)
}

func (c *ParticipantBankController) getParticipantBanks(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodGet {
		response := commons.ErrorResponse[[]models.ParticipantBankResponse]("method not allowed")
		c.respondError(w, http.StatusMethodNotAllowed, response, r, start)
		return
	}

	logRequest(r, nil)
	response, err := c.service.GetParticipantBanks(r.Context())
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		c.respondError(w, http.StatusInternalServerError, response, r, start)
		return
	}

	c.respondSuccess(w, http.StatusOK, response, r, start)
}

// respondSuccess sends a successful JSON response with logging
func (c *ParticipantBankController) respondSuccess(w http.ResponseWriter, status int, payload any, r *http.Request, start time.Time) {
	if err := writeJSON(w, status, payload); err != nil {
		logError(r, err, logger.Fields{"action": "write response"})
	}
	logResponse(r, status, payload, start)
}

// respondError sends an error JSON response with logging
func (c *ParticipantBankController) respondError(w http.ResponseWriter, status int, payload any, r *http.Request, start time.Time) {
	if err := writeJSON(w, status, payload); err != nil {
		logError(r, err, logger.Fields{"action": "write error response"})
	}
	logResponse(r, status, payload, start)
}
