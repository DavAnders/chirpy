package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	var newChirp Chirp
	if err := json.NewDecoder(r.Body).Decode(&newChirp); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	createdChirp, err := cfg.database.CreateChirp(newChirp.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, createdChirp) // http.StatusCreated = 201
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.database.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve chirps")
		return
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func cleanChirpContent(chirp string) string {
	profaneWords := map[string]string{
		"kerfuffle": "****",
		"sharbert":  "****",
		"fornax":    "****",
	}
	words := strings.Fields(chirp)
	for i, word := range words {
		if replacement, ok := profaneWords[strings.ToLower(word)]; ok {
			words[i] = replacement
		}
	}
	return strings.Join(words, " ")
}

func (cfg *apiConfig) handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		ID   int    `json:"id,omitempty"`
		Body string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanedContent := cleanChirpContent(params.Body)

	chirp, err := cfg.database.CreateChirp(cleanedContent)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to save chirp")
		return
	}

	respondWithJSON(w, http.StatusOK, returnVals{
		ID:   chirp.ID,
		Body: cleanedContent,
	})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}
