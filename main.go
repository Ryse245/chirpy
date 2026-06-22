package main

import (
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func main() {
	apiCfg := apiConfig{fileServerHits: atomic.Int32{}}
	serverMux := http.NewServeMux()
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	serverMux.Handle("/app/", (&apiCfg).middlewareMetricsInc(handler))
	serverMux.HandleFunc("/healthz", healthHandler)
	serverMux.HandleFunc("/metrics", (&apiCfg).metricsHandler)
	serverMux.HandleFunc("/reset", (&apiCfg).resetHandler)
	server := http.Server{Handler: serverMux, Addr: ":8080"}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
}

func healthHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) metricsHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Metrics Hit")
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	str := fmt.Sprintf("Hits: %d\n", cfg.fileServerHits.Load())
	_, err := writer.Write([]byte(str))
	if err != nil {
		fmt.Printf("%v", err)
	}
}

func (cfg *apiConfig) resetHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Reset Hit")
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	cfg.fileServerHits.Store(0)
	str := fmt.Sprintf("Hits: %d", cfg.fileServerHits.Load())
	_, err := writer.Write([]byte(str))
	if err != nil {
		fmt.Printf("%v", err)
	}
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hit Middle")
		cfg.fileServerHits.Store(cfg.fileServerHits.Add(1))
		next.ServeHTTP(w, r)
	})
}
