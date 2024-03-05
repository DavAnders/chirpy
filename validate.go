package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	tokenString, err := GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: No token provided")
		return
	}

	userIDstr, err := ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Can't validate JWT")
		return
	}
	userID, err := strconv.Atoi(userIDstr)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: Invalid token")
		return
	}

	var newChirp struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&newChirp); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	createdChirp, err := cfg.database.CreateChirp(newChirp.Body, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, createdChirp) // http.StatusCreated = 201
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	authorIDParam := r.URL.Query().Get("author_id")
	if authorIDParam != "" {
		authorID, err := strconv.Atoi(authorIDParam)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author ID")
			return
		}

		chirps, err := cfg.database.GetChirpsByAuthorID(authorID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Could not retrieve chirps for the given author")
			return
		}

		respondWithJSON(w, http.StatusOK, chirps)
		return
	}

	chirps, err := cfg.database.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve chirps")
		return
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetChirpsByID(w http.ResponseWriter, r *http.Request) {
	chirpIDString := chi.URLParam(r, "chirpID")
	chirpID, err := strconv.Atoi(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}

	chirp, err := cfg.database.GetChirpByID(chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}

	respondWithJSON(w, http.StatusOK, chirp)
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
	// Extract JWT token from Authorization header
	tokenString, tokenErr := GetBearerToken(r.Header)
	if tokenErr != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization token is required")
		return
	}

	// Validate the JWT and extract user ID
	userIDString, idErr := ValidateJWT(tokenString, cfg.jwtSecret)
	if idErr != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	userID, userErr := strconv.Atoi(userIDString)
	if userErr != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

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

	chirp, err := cfg.database.CreateChirp(cleanedContent, userID)
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
