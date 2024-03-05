package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (db *DB) UpgradeUserToChirpyRed(userID int) error {
	dbStruct, err := db.loadDB()
	if err != nil {
		return err
	}

	user, ok := dbStruct.Users[userID]
	if !ok {
		return fmt.Errorf("user not found")
	}

	user.IsChirpyRed = true
	dbStruct.Users[userID] = user

	return db.writeDB(dbStruct)
}

func (cfg *apiConfig) handlePolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	apiKeyHeader := r.Header.Get("Authorization")
	expectedAuthHeader := "ApiKey " + cfg.polkaAPIKey

	if apiKeyHeader != expectedAuthHeader {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: Invalid API key")
		return
	}

	var webhook struct {
		Event string `json:"event"`
		Data  struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Respond immediately for non-relevant events
	if webhook.Event != "user.upgraded" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Upgrade the user in database
	if err := cfg.database.UpgradeUserToChirpyRed(webhook.Data.UserID); err != nil {
		if err.Error() == "user not found" {
			respondWithError(w, http.StatusNotFound, "User not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to upgrade user")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
