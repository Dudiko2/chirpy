package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/Dudiko2/chirpy/internal/db"
)

func replaceBadWords(s string, words []string) string {
	sep := " "
	parts := strings.Split(s, sep)
	for i, p := range parts {
		lc := strings.ToLower(p)
		index := slices.Index(words, lc)
		if index > -1 {
			parts[i] = "****"
		}
	}
	return strings.Join(parts, sep)
}

func handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Body string `json:"body"`
	}
	badWords := []string{
		"kerfuffle", "sharbert", "fornax",
	}
	decoder := json.NewDecoder(r.Body)
	params := reqBody{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Internal error")
		return
	}
	textLen := len(params.Body)
	if textLen > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	sanitizedBody := replaceBadWords(params.Body, badWords)
	chirp, err := database.CreateChirp(sanitizedBody)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respondWithJSON(w, http.StatusCreated, chirp)
}

func handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := database.GetChirps()
	if err != nil {
		log.Printf("Error getting chirps %v", err)
		respondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respondWithJSON(w, http.StatusOK, chirps)
}

func handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("chirpID")
	chirpID, err := strconv.Atoi(idString)
	if err != nil {
		log.Printf("Error converting chripID string %v", err)
		respondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	chirp, err := database.GetChirp(chirpID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "not found")
			return
		}
		log.Printf("Error getting chirp from DB %v", err)
		respondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respondWithJSON(w, http.StatusOK, chirp)
}
