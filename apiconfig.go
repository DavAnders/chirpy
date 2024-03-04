package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

type apiConfig struct {
	fileserverHits int
	mu             sync.Mutex
	database       *DB
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.mu.Lock()
		cfg.fileserverHits++
		cfg.mu.Unlock()
		//mt.Printf("Fileserver hits: %d\n", cfg.fileserverHits)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templatePath := filepath.Join("templates", "admin.html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	cfg.mu.Lock()
	visits := cfg.fileserverHits
	defer cfg.mu.Unlock()
	err = tmpl.Execute(w, struct{ Visits int }{Visits: visits})
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// fmt.Fprintf(w, "Hits: %d", cfg.fileserverHits)
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.mu.Lock()
	cfg.fileserverHits = 0
	cfg.mu.Unlock()
	w.Write([]byte("Counter reset"))
}
