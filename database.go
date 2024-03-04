package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type DB struct {
	Path string
	Mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
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

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	// generate ID
	newID := len(dbStruct.Chirps) + 1
	newChirp := Chirp{ID: newID, Body: body}
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

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	db.Mux.Lock()
	defer db.Mux.Unlock()
	fmt.Println("ensure")

	if _, err := os.Stat(db.Path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	initialDB := DBStructure{Chirps: make(map[int]Chirp)}
	return db.writeDB(initialDB)
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
	fmt.Println("write")
	bytes, err := json.Marshal(&dbStructure)
	if err != nil {
		return err
	}
	file := os.WriteFile(db.Path, bytes, 0644) // filemode = binary perms
	return file
}
