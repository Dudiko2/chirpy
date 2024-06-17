package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/Dudiko2/chirpy/internal/db"
	"golang.org/x/crypto/bcrypt"
)

type EmailAndPass struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SanitizedUser struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

func SanitizeUser(u db.User) SanitizedUser {
	return SanitizedUser{
		ID:    u.ID,
		Email: u.Email,
	}
}

func handlerPostUser(w http.ResponseWriter, r *http.Request) {
	type reqBody EmailAndPass
	params, err := parseBodyJSON[reqBody](r)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Internal error")
		return
	}
	if len(params.Password) < 6 {
		respondWithError(w, http.StatusBadRequest, "password is too short")
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(params.Password), 10)
	if err != nil {
		log.Printf("Error encrypting password: %s", err)
		respondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	user, err := database.CreateUser(params.Email, string(hashed))
	if err != nil {
		if errors.Is(err, db.ErrEmailTaken) {
			respondWithError(w, http.StatusConflict, "conflict")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respondWithJSON(w, http.StatusCreated, SanitizeUser(user))
}

func handlerLogin(w http.ResponseWriter, r *http.Request) {
	type reqBody EmailAndPass
	params, err := parseBodyJSON[reqBody](r)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Internal error")
		return
	}
	user, err := database.GetUserByEmail(params.Email)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			respondWithError(w, http.StatusBadRequest, "Bad request")
			return
		}
		log.Printf("Error getting user %v", err)
		respondWithError(w, http.StatusInternalServerError, "Internal error")
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password),
		[]byte(params.Password))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	respondWithJSON(w, http.StatusOK, SanitizeUser(user))
}
