package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/Dudiko2/chirpy/internal/db"
	"github.com/joho/godotenv"
)

var database *db.DB

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

func parseBodyJSON[V any](r *http.Request) (V, error) {
	decoder := json.NewDecoder(r.Body)
	var params V
	err := decoder.Decode(&params)
	if err != nil {
		var empty V
		return empty, err
	}
	return params, err
}

func main() {
	godotenv.Load()
	var err error
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	dbPath := "database.json"
	if *dbg {
		db.RemoveDB(dbPath)
	}
	database, err = db.NewDB(dbPath)
	if err != nil {
		log.Fatal("Failed to start DB", err)
	}
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
	mux.HandleFunc("POST /api/chirps", handlerPostChirp)
	mux.HandleFunc("GET /api/chirps", handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", handlerGetChirp)
	mux.HandleFunc("POST /api/users", handlerPostUser)
	mux.HandleFunc("POST /api/login", handlerLogin)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	server := http.Server{
		Addr:    "localhost:" + port,
		Handler: mux,
	}
	log.Printf("Listening on port: %s\n", port)
	server.ListenAndServe()
}
