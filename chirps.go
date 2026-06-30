package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hardiing/chirpy/internal/auth"
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

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		msg := fmt.Sprintf("Error getting authentication header: %s", err)
		respondWithError(w, 500, msg)
		return
	}
	validatedUser, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		msg := fmt.Sprintf("Error validating JWT: %s", err)
		respondWithError(w, 401, msg)
		return
	}

	query_params := database.CreateChirpParams{
		Body:   checkedPost,
		UserID: validatedUser,
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

func (cfg *apiConfig) getSingleChirpHandler(w http.ResponseWriter, r *http.Request) {
	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		User_ID   uuid.UUID `json:"user_id"`
	}

	path := r.PathValue("chirpID")
	parsedUUID, err := uuid.Parse(path)
	if err != nil {
		msg := fmt.Sprintf("Invalid UUID string: %s", err)
		respondWithError(w, 500, msg)
	}

	chirp, err := cfg.db.GetSingleChirp(r.Context(), parsedUUID)
	convertChirp := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		User_ID:   chirp.UserID,
	}
	if err != nil {
		msg := fmt.Sprintf("Error getting single chirp: %s", err)
		respondWithError(w, 404, msg)
		return
	}

	respondWithJSON(w, 200, convertChirp)
}

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		msg := fmt.Sprintf("Error getting token: %s", err)
		respondWithError(w, 401, msg)
		return
	}

	validated_user, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		msg := fmt.Sprintf("Error validating JWT: %s", err)
		respondWithError(w, 401, msg)
		return
	}

	path := r.PathValue("chirpID")
	parsedUUID, err := uuid.Parse(path)
	if err != nil {
		msg := fmt.Sprintf("Invalid UUID string: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	chirp, err := cfg.db.GetSingleChirp(r.Context(), parsedUUID)
	if err != nil {
		msg := fmt.Sprintf("Error getting chirp: %s", err)
		respondWithError(w, 404, msg)
		return
	}

	if chirp.UserID != validated_user {
		msg := "You do not own this chirp"
		respondWithError(w, 403, msg)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), parsedUUID)
	if err != nil {
		msg := fmt.Sprintf("Error deleting chirp: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
