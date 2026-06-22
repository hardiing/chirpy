package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

func validateHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		Cleaned string `json:"cleaned_body"`
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

	respondWithJSON(w, 200, returnVals{Cleaned: checkedPost})
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
