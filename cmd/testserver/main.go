// Command testserver serves the repository so you can open fx.test.html in a
// real browser and watch the suite run.
//
// `go test ./...` does this for you in headless Chrome. This is for the times
// when a test is failing and you want to be in there with a debugger.
//
//	go run ./cmd/testserver
//	open http://localhost:8081/fx.test.html
package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))

	// The page beacons its results here. Print them instead of ignoring them,
	// so the terminal is useful even when the browser is doing the work.
	mux.HandleFunc("/__fx_results", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		log.Printf("results: %s", body)
		w.WriteHeader(http.StatusNoContent)
	})

	log.Printf("open http://localhost:%s/fx.test.html", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
