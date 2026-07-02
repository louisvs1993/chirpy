package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
)

//
// API CONFIG FOR SERVERHITS
//

type apiConfig struct {
	fileserverHits atomic.Int32
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
    	Valid bool `json:"valid"`
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

	respBody := returnVals{
		Valid: true,
	}

	respondWithJSON(w, http.StatusOK, respBody)
}

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
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
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