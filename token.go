package main

import (
	"net/http"
	"time"
)

type Revocation struct {
	Token     string    `json:"token"`
	RevokedAt time.Time `json:"revoked_at"`
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	// Use the utility function to extract the refresh token from the Authorization header
	refreshToken, err := GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't extract token from Authorization header")
		return
	}

	// Revoke the token using database method
	err = cfg.database.RevokeToken(refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to revoke token")
		return
	}

	// Respond to indicate successful revocation
	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Token revoked successfully",
	})
}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	// Extract the token using the utility function
	refreshToken, err := GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if the token is revoked
	isRevoked, err := cfg.database.isTokenRevoked(refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to check token revocation status")
		return
	}
	if isRevoked {
		respondWithError(w, http.StatusUnauthorized, "Token has been revoked")
		return
	}

	// Generate a new access token
	newAccessToken, err := RefreshToken(refreshToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to refresh token")
		return
	}

	// Respond with the new access token
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"token": newAccessToken,
	})
}

func (db *DB) RevokeToken(tokenString string) error {
	dbStruct, err := db.loadDB()
	if err != nil {
		return err
	}

	dbStruct.Revocations[tokenString] = Revocation{
		Token:     tokenString,
		RevokedAt: time.Now().UTC(),
	}
	return db.writeDB(dbStruct)
}

func (db *DB) isTokenRevoked(tokenString string) (bool, error) {

	dbStruct, err := db.loadDB()
	if err != nil {
		return false, err
	}

	revocation, ok := dbStruct.Revocations[tokenString]
	if !ok {
		return false, nil
	}
	return !revocation.RevokedAt.IsZero(), nil
}
