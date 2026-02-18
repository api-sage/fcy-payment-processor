package router

import "net/http"

type AccountRouteRegistrar interface {
	RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler)
}

type UserRouteRegistrar interface {
	RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler)
}

func New(accountController AccountRouteRegistrar, userController UserRouteRegistrar, authMiddleware func(http.Handler) http.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	registerSwaggerRoutes(mux)
	mux.Handle("/verify-user-pin", http.RedirectHandler("/verify-pin", http.StatusMovedPermanently))

	if accountController != nil {
		accountController.RegisterRoutes(mux, authMiddleware)
	}
	if userController != nil {
		userController.RegisterRoutes(mux, authMiddleware)
	}

	return mux
}
