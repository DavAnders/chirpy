package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Password string `json:"password,omitempty"`
	Email    string `json:"email"`
	ID       int    `json:"id"`
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	respondWithJSON(w, http.StatusOK, User{ID: user.ID, Email: user.Email})
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
