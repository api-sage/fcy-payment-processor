package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/app"
)

func main() {
	startupStart := time.Now()

	handler, err := app.New(context.Background())
	if err != nil {
		log.Fatalf("%v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	startupDuration := time.Since(startupStart)
	log.Printf("==================================================")
	log.Printf("Application startup completed")
	log.Printf("Start time: %v", startupStart.Format("2006-01-02 15:04:05.000"))
	log.Printf("End time:   %v", time.Now().Format("2006-01-02 15:04:05.000"))
	log.Printf("Total time: %v", startupDuration)
	log.Printf("==================================================")
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("start http server: %v", err)
	}
}
