// Command server runs fx.ciuffolo.com.
//
// It is the same binary locally and in production: Vercel's Go preset builds
// this package and runs it behind its proxy, so there is no serverless handler
// shim and nothing that only exists in one environment.
//
//	go run ./cmd/server
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kilianc/fragment-exchange/web"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           web.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("fx: listening on http://localhost:%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
