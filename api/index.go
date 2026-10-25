package handler

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/api-sage/fcy-payment-processor/src/app"
)

var (
	mu      sync.Mutex
	handler http.Handler
)

// Handler is the Vercel Go entrypoint. The app is built on the first request
// and reused by warm invocations. A failed build is retried on the next request.
func Handler(w http.ResponseWriter, r *http.Request) {
	h, err := get()
	if err != nil {
		log.Printf("app init failed: %v", err)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	h.ServeHTTP(w, r)
}

func get() (http.Handler, error) {
	mu.Lock()
	defer mu.Unlock()

	if handler != nil {
		return handler, nil
	}

	h, err := app.New(context.Background())
	if err != nil {
		return nil, err
	}

	handler = h
	return handler, nil
}
