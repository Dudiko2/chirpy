package main

import (
	"log"
	"net/http"
)

func main() {
	rootDirPath := http.Dir(".")
	appPath := "/app/"
	port := "8080"
	mux := http.NewServeMux()
	mux.Handle(appPath+"*",
		http.StripPrefix(appPath, http.FileServer(rootDirPath)))
	mux.HandleFunc("/healthz", handlerHealth)
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
