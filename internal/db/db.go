package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

var ErrNotFound = fmt.Errorf("not found")

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
	db := &DB{path: path, mux: &sync.RWMutex{}}
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
	for {
		_, taken := data.Chirps[newID]
		if !taken {
			break
		}
		newID++
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

func (db *DB) GetChirps() ([]Chirp, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return []Chirp{}, err
	}
	res := make([]Chirp, 0, len(dbStruct.Chirps))
	for _, c := range dbStruct.Chirps {
		res = append(res, c)
	}
	return res, nil
}

func (db *DB) GetChirp(id int) (Chirp, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	c, ok := dbStruct.Chirps[id]
	if !ok {
		return Chirp{}, ErrNotFound
	}
	return c, nil
}

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
	db.mux.RLock()
	defer db.mux.RUnlock()
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
	db.mux.Lock()
	defer db.mux.Unlock()
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
