package main

import (
	"encoding/json"
	"net/http"

	"github.com/louisvs1993/chirpy/internal/auth"
)


func (cfg *apiConfig) handlerUsersLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}
	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}



	user, err := cfg.db.GetUsersByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get user", err)
		return
	}

	isPasswordCorrect, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't compare passwords.", err)
		return
	}

	if isPasswordCorrect{
		respondWithJSON(w, http.StatusOK, response{
			User: User{
				ID:        user.ID,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				Email:     user.Email,
			},
		})
	} else {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password.", err)
		return
	}
}
