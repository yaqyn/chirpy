package main

import (
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/yaqyn/chirpy/internal/auth"
	"io"
	"net/http"
)

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	key, err := auth.GetAPIKey(r.Header)
	if err != nil || subtle.ConstantTimeCompare([]byte(key), []byte(cfg.polkaKey)) != 1 {
		respondWithError(w, http.StatusUnauthorized, "Invalid or missing API key")
		return
	}
	var params struct {
		Event string          `json:"event"`
		Data  json.RawMessage `json:"data"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	var data struct {
		UserID uuid.UUID `json:"user_id"`
	}
	if err := json.Unmarshal(params.Data, &data); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user data")
		return
	}
	_, err = cfg.dbQueries.UpgradeUserToChirpyRed(r.Context(), data.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to upgrade user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
