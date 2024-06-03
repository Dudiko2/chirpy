package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
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
