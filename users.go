package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hardiing/chirpy/internal/auth"
	"github.com/hardiing/chirpy/internal/database"
)

func (cfg *apiConfig) usersHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type User struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		msg := fmt.Sprintf("Error decoding parameters: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	hashedPW, err := auth.HashPassword(params.Password)
	if err != nil {
		msg := fmt.Sprintf("Error hashing password: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	query_params := database.CreateUserParams{
		HashedPassword: hashedPW,
		Email:          params.Email,
	}

	createdUser, err := cfg.db.CreateUser(r.Context(), query_params)
	if err != nil {
		msg := fmt.Sprintf("Error creating user: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	apiUser := User{
		ID:          createdUser.ID,
		CreatedAt:   createdUser.CreatedAt,
		UpdatedAt:   createdUser.UpdatedAt,
		Email:       createdUser.Email,
		IsChirpyRed: createdUser.IsChirpyRed,
	}

	respondWithJSON(w, 201, apiUser)
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type User struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		msg := fmt.Sprintf("Error decoding parameters: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	user, err := cfg.db.UserLookup(r.Context(), params.Email)
	if err != nil {
		msg := "Incorrect email or password"
		respondWithError(w, 401, msg)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		msg := "Incorrect email or password"
		respondWithError(w, 401, msg)
		return
	}
	if match == false {
		msg := "Incorrect email or password"
		respondWithError(w, 401, msg)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret)
	if err != nil {
		msg := fmt.Sprintf("Error making JWT: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	refresh_token := auth.MakeRefreshToken()

	query_params := database.CreateRefreshTokenParams{
		Token:     refresh_token,
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	}
	refresh_db, err := cfg.db.CreateRefreshToken(r.Context(), query_params)
	if err != nil {
		msg := fmt.Sprintf("Error creating refresh token: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	apiUser := User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        token,
		RefreshToken: refresh_db.Token,
	}

	respondWithJSON(w, 200, apiUser)
}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	type Token struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		msg := fmt.Sprintf("Error getting token: %s", err)
		respondWithError(w, 500, msg)
		return
	}
	db_token, err := cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		msg := fmt.Sprintf("Error getting refresh token from db: %s", err)
		respondWithError(w, 401, msg)
		return
	}

	if time.Now().After(db_token.ExpiresAt) || db_token.RevokedAt.Valid {
		msg := "Token has either expired or is revoked"
		respondWithError(w, 401, msg)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), db_token.Token)
	if err != nil {
		msg := fmt.Sprintf("Error getting user from refresh token: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	newAccessToken, err := auth.MakeJWT(user, cfg.secret)
	if err != nil {
		msg := fmt.Sprintf("Error creating JWT: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	refreshedToken := Token{
		Token: newAccessToken,
	}

	respondWithJSON(w, 200, refreshedToken)
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		msg := fmt.Sprintf("Error getting token: %s", err)
		respondWithError(w, 500, msg)
		return
	}
	db_token, err := cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		msg := fmt.Sprintf("Error getting refresh token from db: %s", err)
		respondWithError(w, 401, msg)
		return
	}

	if time.Now().After(db_token.ExpiresAt) || db_token.RevokedAt.Valid {
		msg := "Token has either expired or is revoked"
		respondWithError(w, 401, msg)
		return
	}

	err = cfg.db.RevokeToken(r.Context(), db_token.Token)
	if err != nil {
		msg := fmt.Sprintf("Error revoking token: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) updateHandler(w http.ResponseWriter, r *http.Request) {
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

	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type User struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		msg := fmt.Sprintf("Error decoding parameters: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	hashedPW, err := auth.HashPassword(params.Password)
	if err != nil {
		msg := fmt.Sprintf("Error hashing password: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	query_params := database.UpdateUserParams{
		HashedPassword: hashedPW,
		Email:          params.Email,
		ID:             validated_user,
	}

	updatedUser, err := cfg.db.UpdateUser(r.Context(), query_params)
	if err != nil {
		msg := fmt.Sprintf("Error updating user: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	apiUser := User{
		ID:          updatedUser.ID,
		CreatedAt:   updatedUser.CreatedAt,
		UpdatedAt:   updatedUser.UpdatedAt,
		Email:       updatedUser.Email,
		IsChirpyRed: updatedUser.IsChirpyRed,
	}

	respondWithJSON(w, 200, apiUser)
}
