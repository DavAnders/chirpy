package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	// Extract chirpID
	chirpIDstr := chi.URLParam(r, "chirpID")
	chirpID, err := strconv.Atoi(chirpIDstr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}

	// Extract JWT token
	tokenString, err := GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization token is required")
		return
	}

	// Validate JWT and extract user ID
	userIDString, err := ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	userID, err := strconv.Atoi(userIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Retrieve chirp to check if user is author
	chirp, err := cfg.database.GetChirpByID(chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}

	// Check if user is author
	if chirp.AuthorID != userID {
		respondWithError(w, http.StatusForbidden, "User is not the author of the chirp")
		return
	}

	// Delete chirp
	err = cfg.database.DeleteChirp(chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete chirp")
		return
	}

	w.WriteHeader(http.StatusOK)
}
