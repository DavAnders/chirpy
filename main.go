package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter() // Chi router
	apiRouter := chi.NewRouter()
	admin := chi.NewRouter()
	cfg := &apiConfig{} // apiconfig
	const port = "8080"

	fileServer := http.FileServer(http.Dir(".")) // project root
	wrappedFileServer := cfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer))
	// Chi handles routing now
	r.Handle("/app/*", wrappedFileServer)
	r.Handle("/app", wrappedFileServer)

	apiRouter.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	admin.Get("/metrics", cfg.metricsHandler) // only GET
	apiRouter.HandleFunc("/reset", cfg.resetHandler)

	// mount before server config
	r.Mount("/api", apiRouter)
	r.Mount("/admin", admin)

	corsHandler := middlewareCors(r)
	server := &http.Server{
		Handler: corsHandler,
		Addr:    "localhost:" + port,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
