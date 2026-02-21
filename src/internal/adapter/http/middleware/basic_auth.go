package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
)

func BasicAuth(channelID, channelKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if channelID == "" || channelKey == "" {
				logger.Error("basic auth middleware missing server configuration", nil, logger.Fields{
					"method": r.Method,
					"path":   r.URL.Path,
				})
				http.Error(w, "server auth configuration is missing", http.StatusInternalServerError)
				return
			}

			id, key, ok := r.BasicAuth()
			if !ok || !secureEqual(id, channelID) || !secureEqual(key, channelKey) {
				logger.Info("basic auth middleware unauthorized request", logger.Fields{
					"method":      r.Method,
					"path":        r.URL.Path,
					"credentials": "invalid_or_missing",
				})
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			logger.Info("basic auth middleware authorized request", logger.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
			})
			next.ServeHTTP(w, r)
		})
	}
}

func secureEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
