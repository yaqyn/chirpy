package main

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/yaqyn/chirpy/internal/auth"
	"github.com/yaqyn/chirpy/internal/database"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"
)

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or missing token")
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or missing token")
		return
	}
	var params struct {
		Body string `json:"body"`
	}
	if !decodeJSON(w, r, &params) {
		return
	}
	if strings.TrimSpace(params.Body) == "" {
		respondWithError(w, http.StatusBadRequest, "Chirp body is required")
		return
	}
	if utf8.RuneCountInString(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body: cleanChirp(params.Body), UserID: userID,
	})
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			respondWithError(w, http.StatusBadRequest, "User does not exist")
			return
		}
		log.Printf("Failed to create chirp: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create chirp")
		return
	}
	respondWithJSON(w, http.StatusCreated, Chirp{
		ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body, UserID: chirp.UserID,
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	params, err := parseChirpQuery(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	rows, err := cfg.dbQueries.GetChirpsPage(r.Context(), params)
	if err != nil {
		log.Printf("Failed to get chirps: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to get chirps")
		return
	}
	chirps := make([]Chirp, 0, len(rows))
	for _, row := range rows {
		chirps = append(chirps, Chirp{
			ID: row.ID, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
			Body: row.Body, UserID: row.UserID,
		})
	}
	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}
	chirp, err := cfg.dbQueries.GetChirp(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}
	if err != nil {
		log.Printf("Failed to get chirp: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to get chirp")
		return
	}
	respondWithJSON(w, http.StatusOK, Chirp{
		ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body, UserID: chirp.UserID,
	})
}

func cleanChirp(body string) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		switch strings.ToLower(word) {
		case "kerfuffle", "sharbert", "fornax":
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or missing token")
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or missing token")
		return
	}
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}
	chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpID)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get chirp")
		return
	}
	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You can only delete your own chirps")
		return
	}
	deleted, err := cfg.dbQueries.DeleteChirp(r.Context(), database.DeleteChirpParams{
		ID: chirpID, UserID: userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete chirp")
		return
	}
	if deleted == 0 {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
