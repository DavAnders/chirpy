package main

import (
	"encoding/json"
	"net/http"
)

type User struct {
	Email string `json:"email"`
	ID    int    `json:"id"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	var newUserRequest struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&newUserRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	newUser, err := cfg.database.CreateUser(newUserRequest.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	respondWithJSON(w, http.StatusCreated, newUser)

}
