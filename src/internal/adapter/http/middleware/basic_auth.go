package middleware

import (
	"crypto/subtle"
	"net/http"
)

const (
	defaultChannelID  = "GreyApp"
	defaultChannelKey = "GrehoundKey001"
)

func BasicAuth(channelID, channelKey string) func(http.Handler) http.Handler {
	if channelID == "" {
		channelID = defaultChannelID
	}
	if channelKey == "" {
		channelKey = defaultChannelKey
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, key, ok := r.BasicAuth()
			if !ok || !secureEqual(id, channelID) || !secureEqual(key, channelKey) {
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
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
