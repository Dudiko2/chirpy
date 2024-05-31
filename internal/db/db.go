package db

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

type Chirp struct {
	ID   uint   `json:"id"`
	Body string `json:"body"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{path: path}
	err := db.ensureDB()
	if err != nil {
		return &DB{}, err
	}
	return db, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	data, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	newID := 1
	for _, taken := data.Chirps[newID]; taken; newID++ {
	}
	newChirp := Chirp{
		ID:   uint(newID),
		Body: body,
	}
	data.Chirps[newID] = newChirp
	writeErr := db.writeDB(data)
	if writeErr != nil {
		return Chirp{}, writeErr
	}
	return newChirp, nil
}

// func (db *DB) GetChirps() ([]Chirp, error) {}

func (db *DB) ensureDB() error {
	_, err := os.Stat(db.path)
	if errors.Is(err, os.ErrNotExist) {
		data, err := json.Marshal(NewDBStructure())
		if err != nil {
			return err
		}
		err = os.WriteFile(db.path, data, 0766)
		if err != nil {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) loadDB() (DBStructure, error) {
	data, readErr := os.ReadFile(db.path)
	if readErr != nil {
		return DBStructure{}, readErr
	}
	res := DBStructure{}
	jsonErr := json.Unmarshal(data, &res)
	if jsonErr != nil {
		return DBStructure{}, jsonErr
	}
	return res, nil
}

func (db *DB) writeDB(dbStruct DBStructure) error {
	data, err := json.Marshal(dbStruct)
	if err != nil {
		return err
	}
	err = os.WriteFile(db.path, data, 0766)
	if err != nil {
		return err
	}
	return nil
}

func NewDBStructure() DBStructure {
	return DBStructure{
		Chirps: map[int]Chirp{},
	}
}
