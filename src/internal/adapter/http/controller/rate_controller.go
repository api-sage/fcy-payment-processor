package controller

import (
	"context"
	"net/http"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
)

type RateService interface {
	GetRates(ctx context.Context) (models.Response[[]models.RateResponse], error)
	GetRate(ctx context.Context, req models.GetRateRequest) (models.Response[models.RateResponse], error)
}

type RateController struct {
	service RateService
}

func NewRateController(service RateService) *RateController {
	return &RateController{service: service}
}

func (c *RateController) RegisterRoutes(_ *http.ServeMux, _ func(http.Handler) http.Handler) {}
