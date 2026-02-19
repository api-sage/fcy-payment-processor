package router

import "net/http"

type AccountRouteRegistrar interface {
	RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler)
}

type UserRouteRegistrar interface {
	RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler)
}

type ParticipantBankRouteRegistrar interface {
	RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler)
}

type RateRouteRegistrar interface {
	RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler)
}

func New(
	accountController AccountRouteRegistrar,
	userController UserRouteRegistrar,
	participantBankController ParticipantBankRouteRegistrar,
	rateController RateRouteRegistrar,
	authMiddleware func(http.Handler) http.Handler,
) *http.ServeMux {
	mux := http.NewServeMux()
	registerSwaggerRoutes(mux)

	if accountController != nil {
		accountController.RegisterRoutes(mux, authMiddleware)
	}
	if userController != nil {
		userController.RegisterRoutes(mux, authMiddleware)
	}
	if participantBankController != nil {
		participantBankController.RegisterRoutes(mux, authMiddleware)
	}
	if rateController != nil {
		rateController.RegisterRoutes(mux, authMiddleware)
	}

	return mux
}
