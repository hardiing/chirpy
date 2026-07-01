package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hardiing/chirpy/internal/auth"
)

func (cfg *apiConfig) webhookHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		msg := fmt.Sprintf("Error decoding parameters: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		msg := fmt.Sprintf("Error getting API Key: %s", err)
		respondWithError(w, 401, msg)
		return
	}

	if apiKey != cfg.api_key {
		respondWithError(w, 401, "Invalid API Key")
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	_, err = cfg.db.UpgradeUser(r.Context(), params.Data.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, 404, "User not found")
			return
		} else {
			respondWithError(w, 500, "Error upgrading user")
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
