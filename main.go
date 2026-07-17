package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/louisvs1993/chirpy/internal/database"
)

//
// API CONFIG FOR SERVERHITS
//

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

//
// API READINESS
//

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

//
// API CALLS
//

func (cfg *apiConfig) handlerNumOfReqs(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerResetNumOfReqs(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Hits: " + strconv.FormatInt(int64(cfg.fileserverHits.Load()), 10)))
}

//
// VALIDATION SECTION
//

func handlerChirpValidation(w http.ResponseWriter, r *http.Request) {
    type parameters struct {
        Body string `json:"body"`
    }

	type returnVals struct {
    	CleanedBody string `json:"cleaned_body"`
	}

    decoder := json.NewDecoder(r.Body)
    params := parameters{}
    err := decoder.Decode(&params)
    if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusBadRequest, "Something went wrong", err)
		return
    }

	if len(params.Body) > 140 {
		log.Printf("Chirp is too long")
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleaned := stringCleaner(params.Body)

	respBody := returnVals{
		CleanedBody: cleaned,
	}

	respondWithJSON(w, http.StatusOK, respBody)
}

func stringCleaner(s string) string{
	words := strings.Split(s, " ")
	bannedWords := []string{"kerfuffle", "sharbert", "fornax"}

    for i, word := range words {
        lower := strings.ToLower(word)

        if slices.Contains(bannedWords, lower){
			words[i] = "****"
		}
    }

    return strings.Join(words, " ")
}

//
// RESPONSE CREATORS
//

func respondWithError(w http.ResponseWriter, code int, msg string, err error){
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}){
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	dat, err := json.Marshal(payload)
	if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
	}
	w.WriteHeader(code)
    w.Write(dat)
}

//
// MAIN FUNCTION
//

func main() {
	const port = "8080"
	const filepathRoot = "."

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConn)
	
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db: dbQueries,
	}
    mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerNumOfReqs)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerResetNumOfReqs)
	mux.HandleFunc("POST /api/validate_chirp", handlerChirpValidation)

	srv := &http.Server{
        Addr:    ":" + port,
        Handler: mux,
    }

    log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}