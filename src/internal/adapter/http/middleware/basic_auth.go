package middleware

import (
	"crypto/subtle"
	"net/http"
)

func BasicAuth(channelID, channelKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if channelID == "" || channelKey == "" {
				http.Error(w, "server auth configuration is missing", http.StatusInternalServerError)
				return
			}

			id, key, ok := r.BasicAuth()
			if !ok || !secureEqual(id, channelID) || !secureEqual(key, channelKey) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func secureEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
