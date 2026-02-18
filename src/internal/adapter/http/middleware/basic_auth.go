package middleware

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
)

const (
	defaultChannelID  = "GreyApp"
	defaultChannelKey = "GrehoundKey001"
)

func BasicAuth(channelID, channelKey string) func(http.Handler) http.Handler {
	return ChannelAuth(channelID, channelKey)
}

func ChannelAuth(channelID, channelKey string) func(http.Handler) http.Handler {
	if channelID == "" {
		channelID = defaultChannelID
	}
	if channelKey == "" {
		channelKey = defaultChannelKey
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Channel-ID")
			key := r.Header.Get("X-Channel-Key")
			if !secureEqual(id, channelID) || !secureEqual(key, channelKey) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(models.ErrorResponse[any]("unauthorized", "invalid channel credentials"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func secureEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
