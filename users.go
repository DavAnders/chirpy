package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Password    string `json:"password,omitempty"`
	Email       string `json:"email"`
	ID          int    `json:"id"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	var loginRequest struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := cfg.database.GetUserByEmail(loginRequest.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid email or password")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	expiresIn := 24 * time.Hour // Default to 24 hours
	if loginRequest.ExpiresInSeconds > 0 {
		requestedExpiresIn := time.Duration(loginRequest.ExpiresInSeconds) * time.Second
		if requestedExpiresIn > expiresIn {
			requestedExpiresIn = expiresIn // Cap at 24 hours
		}
		expiresIn = requestedExpiresIn
	}

	// Prepare claims
	accessClaims := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   strconv.Itoa(user.ID),
	}

	// Create token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	tokenString, err := accessToken.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to sign access token")
		return
	}

	// refresh tokens
	refreshTokenClaims := jwt.RegisteredClaims{
		Issuer:    "chirpy-refresh",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(60 * 24 * time.Hour)),
		Subject:   strconv.Itoa(user.ID),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to sign refresh token")
	}

	// Respond with token and user info
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":            user.ID,
		"email":         user.Email,
		"is_chirpy_red": user.IsChirpyRed,
		"token":         tokenString,
		"refresh_token": refreshTokenString,
	})

}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	var newUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&newUserRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	newUser, err := cfg.database.CreateUser(newUserRequest.Email, newUserRequest.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	respondWithJSON(w, http.StatusCreated, User{ID: newUser.ID, Email: newUser.Email})

}

func (db *DB) GetUserByEmail(email string) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, user := range dbStruct.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return User{}, fmt.Errorf("user not found")
}

func (cfg *apiConfig) handleUpdateUsers(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "Authorization header needed")
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader { // no bearer prefix
		respondWithError(w, http.StatusUnauthorized, "Invalid authorization format")
		return
	}

	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.jwtSecret), nil
	})

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	if claims.Issuer != "chirpy-access" {
		respondWithError(w, http.StatusUnauthorized, "Invalid token type")
		return
	}

	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error parsing User ID")
		return
	}

	var updateReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		respondWithError(w, http.StatusBadRequest, "Error decoding request body")
		return
	}

	var hashedPassword string
	if updateReq.Password != "" {
		bytes, err := bcrypt.GenerateFromPassword([]byte(updateReq.Password), bcrypt.DefaultCost)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error hashing password")
			return
		}
		hashedPassword = string(bytes)
	}

	err = cfg.database.UpdateUser(userID, updateReq.Email, hashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":    userID,
		"email": updateReq.Email,
	})
}
