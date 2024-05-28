package main

import (
	"fmt"
	"log"
	"net/http"
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
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	hits := fmt.Sprintf("Hits: %v", cfg.fileserverHits)
	w.Write([]byte(hits))
}

func (cfg *apiConfig) handlerResetMetrics(w http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(200)
	w.Write([]byte("OK"))
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
	mux.HandleFunc("GET /api/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("GET /api/reset", apiCfg.handlerResetMetrics)
	mux.HandleFunc("GET /api/healthz", handlerHealth)
	server := http.Server{
		Addr:    "localhost:" + port,
		Handler: mux,
	}
	log.Printf("Listening on port: %s\n", port)
	server.ListenAndServe()
}

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
