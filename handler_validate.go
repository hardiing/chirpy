package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func validateHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		Valid bool `json:"valid"`
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

	respondWithJSON(w, 200, returnVals{Valid: true})
}
