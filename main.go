package main

import (
	"log"
	"net/http"
)

func main() {
	cfg := &apiConfig{}
	const port = "8080"
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("."))
	wrappedFileServer := cfg.middlewareMetricsInc(fileServer)
	mux.Handle("/app/", http.StripPrefix("/app/", wrappedFileServer))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.Handle("/files/", wrappedFileServer)
	mux.HandleFunc("/metrics/", cfg.metricsHandler)
	mux.HandleFunc("/reset/", cfg.resetHandler)

	corsMux := middlewareCors(mux)
	server := &http.Server{
		Handler: corsMux,
		Addr:    "localhost:" + port,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
