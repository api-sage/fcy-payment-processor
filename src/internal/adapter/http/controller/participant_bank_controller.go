package controller

import (
	"context"
	"net/http"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
)

type ParticipantBankService interface {
	GetParticipantBanks(ctx context.Context) (models.Response[[]models.ParticipantBankResponse], error)
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
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse[[]models.ParticipantBankResponse]("method not allowed"))
		return
	}

	response, err := c.service.GetParticipantBanks(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response)
		return
	}

	writeJSON(w, http.StatusOK, response)
}
