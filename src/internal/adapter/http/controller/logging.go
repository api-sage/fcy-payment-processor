package controller

import (
	"net/http"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
)

func logRequest(r *http.Request, payload any) {
	logger.Info("http request", logger.Fields{
		"method":  r.Method,
		"path":    r.URL.Path,
		"query":   r.URL.RawQuery,
		"payload": logger.SanitizePayload(payload),
	})
}

func logResponse(r *http.Request, status int, payload any, start time.Time) {
	logger.Info("http response", logger.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"status":     status,
		"durationMs": time.Since(start).Milliseconds(),
		"response":   logger.SanitizePayload(payload),
	})
}

func logError(r *http.Request, err error, extra logger.Fields) {
	fields := logger.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
		"query":  r.URL.RawQuery,
	}
	for k, v := range extra {
		fields[k] = v
	}
	logger.Error("http handler error", err, fields)
}
