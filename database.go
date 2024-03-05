package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type Chirp struct {
	AuthorID int    `json:"author_id"`
	Body     string `json:"body"`
	ID       int    `json:"id"`
}

type DB struct {
	Path string
	Mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps      map[int]Chirp         `json:"chirps"`
	Users       map[int]User          `json:"users"`
	Revocations map[string]Revocation `json:"revocations"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	db := &DB{
		Path: path,
		Mux:  &sync.RWMutex{},
	}
	if err := db.ensureDB(); err != nil {
		return nil, err
	}
	return db, nil
}

// Create new user in DB
func (db *DB) CreateUser(email, password string) (User, error) {

	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, user := range dbStruct.Users {
		if user.Email == email {
			return User{}, fmt.Errorf("email already in use")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	newID := len(dbStruct.Users) + 1 // might need to change this implementation
	newUser := User{ID: newID, Email: email, Password: string(hashedPassword)}
	dbStruct.Users[newID] = newUser

	if err := db.writeDB(dbStruct); err != nil {
		return User{}, err
	}
	return newUser, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string, userID int) (Chirp, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	// generate ID
	newID := len(dbStruct.Chirps) + 1
	newChirp := Chirp{ID: newID, Body: body, AuthorID: userID}
	dbStruct.Chirps[newID] = newChirp

	if err := db.writeDB(dbStruct); err != nil {
		return Chirp{}, err
	}

	return newChirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0, len(dbStruct.Chirps))
	for _, chirp := range dbStruct.Chirps {
		chirps = append(chirps, chirp)
	}
	return chirps, nil
}

func (db *DB) GetChirpByID(id int) (Chirp, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	chirp, ok := dbStruct.Chirps[id]
	if !ok {
		return Chirp{}, fmt.Errorf("chirp not found")
	}

	return chirp, nil
}

func (db *DB) GetChirpsByAuthorID(authorID int) ([]Chirp, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	var filteredChirps []Chirp
	for _, chirp := range dbStruct.Chirps {
		if chirp.AuthorID == authorID {
			filteredChirps = append(filteredChirps, chirp)
		}
	}
	return filteredChirps, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	// locking happens in call to writeDB, so it should not be needed here
	// it looks like it doesn't deadlock, but might mess order of operations
	//db.Mux.Lock()
	//defer db.Mux.Unlock()
	log.Println("Ensuring database exists...")

	if _, err := os.Stat(db.Path); os.IsNotExist(err) {
		initialDB := DBStructure{
			Chirps:      make(map[int]Chirp),
			Users:       make(map[int]User),
			Revocations: make(map[string]Revocation),
		}
		return db.writeDB(initialDB)
	}
	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	db.Mux.RLock()
	defer db.Mux.RUnlock()

	bytes, err := os.ReadFile(db.Path)
	if err != nil {
		return DBStructure{}, err
	}

	var dbStruct DBStructure
	if err := json.Unmarshal(bytes, &dbStruct); err != nil {
		return DBStructure{}, err
	}

	return dbStruct, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	db.Mux.Lock()
	defer db.Mux.Unlock()
	log.Println("Writing to database...")
	bytes, err := json.Marshal(&dbStructure)
	if err != nil {
		log.Printf("Error marshalling database structure: %v\n", err)
		return err
	}
	err = os.WriteFile(db.Path, bytes, 0644) // filemode = binary perms
	if err != nil {
		log.Printf("Error writing database file: %v\n", err)
	} else {
		log.Println("Database written successfully.")
	}
	return err
}

func (db *DB) UpdateUser(id int, email, hashedPassword string) error {
	dbStruct, err := db.loadDB()
	if err != nil {
		return err
	}

	// Check if the user exists
	user, exists := dbStruct.Users[id]
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Update the user's details
	user.Email = email
	if hashedPassword != "" {
		user.Password = hashedPassword
	}
	dbStruct.Users[id] = user

	return db.writeDB(dbStruct)
}

func (db *DB) DeleteChirp(chirpID int) error {
	dbStruct, err := db.loadDB()
	if err != nil {
		return fmt.Errorf("loading database: %v", err)
	}

	// Check if the chirp exists
	_, exists := dbStruct.Chirps[chirpID]
	if !exists {
		return fmt.Errorf("chirp not found")
	}

	// Delete the chirp from the map
	delete(dbStruct.Chirps, chirpID)

	// Write the updated database structure back to the file
	if err := db.writeDB(dbStruct); err != nil {
		return fmt.Errorf("writing database: %v", err)
	}

	return nil
}
