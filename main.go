package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {

	// env loading
	errEnv := godotenv.Load()
	if errEnv != nil {
		log.Fatalf("Error loading .env file")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("JWT_SECRET is not set in .env file")
	}

	_, err := os.ReadFile("database.json")
	if err == nil {
		os.Remove("database.json")
	}
	r := chi.NewRouter() // Chi router
	apiRouter := chi.NewRouter()
	admin := chi.NewRouter()

	dbPath := "database.json"
	log.Println("Initializing database...")
	newDB, err := NewDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized.")

	cfg := &apiConfig{
		database:  newDB,
		jwtSecret: jwtSecret,
	} // apiconfig
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
	apiRouter.Post("/validate_chirp", cfg.handlerChirpsValidate)
	apiRouter.Post("/chirps", cfg.handlerCreateChirp)
	apiRouter.Get("/chirps", cfg.handlerGetChirps)
	apiRouter.Get("/chirps/{chirpID}", cfg.handlerGetChirpsByID)
	apiRouter.Post("/users", cfg.handlerCreateUser)
	apiRouter.Post("/login", cfg.handlerLogin)
	apiRouter.Put("/users", cfg.handleUpdateUsers)

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
