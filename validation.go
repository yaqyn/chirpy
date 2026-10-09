package main

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/yaqyn/threadline-social-api/internal/database"
	"io"
	"net/http"
	"net/mail"
	"strconv"
)

func decodeJSON(w http.ResponseWriter, r *http.Request, payload any) bool {
	decoder := json.NewDecoder(r.Body)
	// Decode the document first so null is never mistaken for an empty object.
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		jsonError(w, err)
		return false
	}
	if string(raw) == "null" {
		respondWithError(w, 400, "JSON object is required")
		return false
	}
	if err := json.Unmarshal(raw, payload); err != nil {
		jsonError(w, err)
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		jsonError(w, err)
		return false
	}
	return true
}
func jsonError(w http.ResponseWriter, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		respondWithError(w, 413, "Request body is too large")
		return
	}
	respondWithError(w, 400, "Invalid JSON")
}
func validEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value && len(value) <= 254
}
func parseChirpQuery(r *http.Request) (database.GetChirpsPageParams, error) {
	p := database.GetChirpsPageParams{PageLimit: 100}
	query := r.URL.Query()
	if value := query.Get("author_id"); value != "" {
		id, err := uuid.Parse(value)
		if err != nil {
			return p, errors.New("Invalid author ID")
		}
		p.AuthorID = uuid.NullUUID{UUID: id, Valid: true}
	}
	switch query.Get("sort") {
	case "", "asc":
	case "desc":
		p.Descending = true
	default:
		return p, errors.New("Sort must be asc or desc")
	}
	for _, item := range []struct {
		name     string
		target   *int32
		min, max int64
	}{
		{"limit", &p.PageLimit, 1, 100}, {"offset", &p.PageOffset, 0, 2147483647},
	} {
		if value := query.Get(item.name); value != "" {
			n, err := strconv.ParseInt(value, 10, 32)
			if err != nil || n < item.min || n > item.max {
				return p, errors.New("Invalid " + item.name)
			}
			*item.target = int32(n)
		}
	}
	return p, nil
}
