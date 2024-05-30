package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
)

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	tmpl := `<html>

	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	
	</html>
	`
	content := fmt.Sprintf(tmpl, cfg.fileserverHits)
	w.Write([]byte(content))
}

func (cfg *apiConfig) handlerResetMetrics(w http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorRes struct {
		Error string `json:"error"`
	}
	res := errorRes{
		Error: msg,
	}
	respondWithJSON(w, code, res)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func replaceBadWords(s string, words []string) string {
	sep := " "
	parts := strings.Split(s, sep)
	for i, p := range parts {
		lc := strings.ToLower(p)
		index := slices.Index(words, lc)
		if index > -1 {
			parts[i] = "****"
		}
	}
	return strings.Join(parts, sep)
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Body string `json:"body"`
	}
	type resBody struct {
		CleanedBody string `json:"cleaned_body"`
	}
	badWords := []string{
		"kerfuffle", "sharbert", "fornax",
	}
	decoder := json.NewDecoder(r.Body)
	params := reqBody{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Internal error")
		return
	}
	textLen := len(params.Body)
	if textLen > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	res := resBody{
		CleanedBody: replaceBadWords(params.Body, badWords),
	}
	respondWithJSON(w, http.StatusOK, res)
}

func main() {
	apiCfg := apiConfig{
		fileserverHits: 0,
	}
	rootDirPath := http.Dir(".")
	appPath := "/app/"
	port := "8080"
	mux := http.NewServeMux()
	fsHandler := http.StripPrefix(appPath,
		apiCfg.middlewareMetricsInc(http.FileServer(rootDirPath)))
	mux.Handle(appPath+"*", fsHandler)
	mux.HandleFunc("GET /api/reset", apiCfg.handlerResetMetrics)
	mux.HandleFunc("GET /api/healthz", handlerHealth)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	server := http.Server{
		Addr:    "localhost:" + port,
		Handler: mux,
	}
	log.Printf("Listening on port: %s\n", port)
	server.ListenAndServe()
}
