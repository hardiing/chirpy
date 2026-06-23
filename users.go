package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) usersHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		msg := fmt.Sprintf("Error decoding parameters: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	createdUser, err := cfg.db.CreateUser(r.Context(), params.Email)
	if err != nil {
		msg := fmt.Sprintf("Error creating user: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	apiUser := User{
		ID:        createdUser.ID,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Email:     createdUser.Email,
	}

	respondWithJSON(w, 201, apiUser)
}
