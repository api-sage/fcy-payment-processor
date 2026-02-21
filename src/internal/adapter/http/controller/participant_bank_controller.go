package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/commons"
	"github.com/api-sage/ccy-payment-processor/src/internal/logger"
)

type ParticipantBankService interface {
	GetParticipantBanks(ctx context.Context) (commons.Response[[]models.ParticipantBankResponse], error)
}

type ParticipantBankController struct {
	service ParticipantBankService
}

func NewParticipantBankController(service ParticipantBankService) *ParticipantBankController {
	return &ParticipantBankController{service: service}
}

func (c *ParticipantBankController) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	handler := http.HandlerFunc(c.getParticipantBanks)
	if authMiddleware != nil {
		handler = authMiddleware(handler).ServeHTTP
	}

	mux.Handle("/get-participant-banks", http.HandlerFunc(handler))
}

func (c *ParticipantBankController) getParticipantBanks(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logRequest(r, nil)

	if r.Method != http.MethodGet {
		response := commons.ErrorResponse[[]models.ParticipantBankResponse]("method not allowed")
		writeJSON(w, http.StatusMethodNotAllowed, response)
		logResponse(r, http.StatusMethodNotAllowed, response, start)
		return
	}

	response, err := c.service.GetParticipantBanks(r.Context())
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		writeJSON(w, http.StatusInternalServerError, response)
		logResponse(r, http.StatusInternalServerError, response, start)
		return
	}

	writeJSON(w, http.StatusOK, response)
	logResponse(r, http.StatusOK, response, start)
}
