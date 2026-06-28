package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
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
	serverMux.HandleFunc("GET /api/healthz", healthHandler)
	serverMux.HandleFunc("GET /admin/metrics", (&apiCfg).metricsHandler)
	serverMux.HandleFunc("POST /admin/reset", (&apiCfg).resetHandler)
	serverMux.HandleFunc("POST /api/validate_chirp", validateHandler)
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

func validateHandler(writer http.ResponseWriter, request *http.Request) {

	type reqJson struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(request.Body)
	params := reqJson{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(writer, 400, "Error in decoding JSON")
		return
	}
	if len(params.Body) > 140 {
		respondWithError(writer, 400, "Chirp is too long")
		return
	}

	type returnVal struct {
		Cleaned_body string `json:"cleaned_body"`
	}

	censorWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {}}

	strSlice := strings.Split(params.Body, " ")
	for i, word := range strSlice {
		lower := strings.ToLower(word)
		if _, ok := censorWords[lower]; ok {
			strSlice[i] = "****"
		}
	}

	cleanedStr := strings.Join(strSlice, " ")

	respondWithJSON(writer, 200, returnVal{Cleaned_body: cleanedStr})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(msg))
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(`{"error":"Error encoding response JSON"}`))
		return
	}
	w.WriteHeader(code)
	w.Write(dat)

}

func (cfg *apiConfig) metricsHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Metrics Hit")
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	str := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>",
		cfg.fileServerHits.Load())
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
