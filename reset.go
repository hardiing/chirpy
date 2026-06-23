package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		msg := fmt.Sprintf("Reset access forbidden")
		respondWithError(w, 403, msg)
		return
	}
	cfg.fileserverHits.Store(0)
	err := cfg.db.ResetUsers(r.Context())
	if err != nil {
		msg := fmt.Sprintf("Error resseting users table: %s: err")
		respondWithError(w, 500, msg)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0, users table reset"))
}
