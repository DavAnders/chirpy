package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app/", http.StripPrefix("/app/", fileServer))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	corsMux := middlewareCors(mux)
	server := &http.Server{
		Handler: corsMux,
		Addr:    "localhost:" + port,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
