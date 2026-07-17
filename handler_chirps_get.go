package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retreive any chirps.", err)
		return
	}

	resp := []Chirp{};

	for _, chirp := range chirps{
		resp = append(resp, Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		})
	}


    respondWithJSON(w, http.StatusOK, resp)
}

func (cfg *apiConfig) handlerChirpsGetById(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Chirp
	}

	idString := r.PathValue("id")

	id, err := uuid.Parse(idString)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
        return
    }

	chirp, err := cfg.db.GetChirpsById(r.Context(), id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            respondWithError(w, http.StatusNotFound, "Chirp does not exist.", err)
            return
        }
        respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirp.", err)
        return
    }

	respondWithJSON(w, http.StatusOK, response{
		Chirp: Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		},
	})
}