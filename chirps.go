package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hardiing/chirpy/internal/database"
)

func (cfg *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body    string    `json:"body"`
		User_ID uuid.UUID `json:"user_id"`
	}

	type returnVals struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		User_ID   uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		msg := fmt.Sprintf("Error decoding parameters: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	const maxChirpLength = 140

	if len(params.Body) > maxChirpLength {
		msg := "Chirp is too long"
		respondWithError(w, 400, msg)
		return
	}

	checkedPost := profaneCheck(params.Body)

	query_params := database.CreateChirpParams{
		Body:   checkedPost,
		UserID: params.User_ID,
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), query_params)
	if err != nil {
		msg := fmt.Sprintf("Error creating chirp: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	respondWithJSON(w, 201, returnVals{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		User_ID:   chirp.UserID,
	})
}

func profaneCheck(chirp string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	post := strings.Fields(chirp)
	for i, word := range post {
		if slices.Contains(badWords, strings.ToLower(word)) {
			post[i] = "****"
		}
	}
	cleanedChirp := strings.Join(post, " ")
	return cleanedChirp
}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		User_ID   uuid.UUID `json:"user_id"`
	}

	var allChirps []Chirp

	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		msg := fmt.Sprintf("Error getting all chirps: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	for _, chirp := range chirps {
		convertChirp := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			User_ID:   chirp.UserID,
		}
		allChirps = append(allChirps, convertChirp)
	}

	respondWithJSON(w, 200, allChirps)
}
