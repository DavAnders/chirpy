package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"
	mux := http.NewServeMux()
	corsMux := middlewareCors(mux)
	server := &http.Server{
		Handler: corsMux,
		Addr:    "localhost:" + port,
	}
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
