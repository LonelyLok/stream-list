package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"example.com/backend/api"
	"github.com/rs/cors"
)

func init() {
	body, err := os.ReadFile("api/.env")
	if err != nil {
		log.Printf("unable to read file: %v", err)
		return
	}
	fileString := string(body)

	lines := strings.Split(fileString, "\n")

	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		key, value := parts[0], parts[1]
		os.Setenv(key, value)
	}
}

func baseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintf(w, "Hello, World.")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintf(w, "OK")
}

func getUpcomingStreamsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	origin := r.Header.Get("Origin")
	fmt.Printf("Request originated from: %s\n", origin)

	results := api.GetAllUpcomingStreamsByScraping()

	// 2) set your no-cache and content headers
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "application/json")

	// 3) write out the JSON
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "failed to encode JSON", http.StatusInternalServerError)
	}
}

func main() {
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"https://*.onrender.com"},
		AllowedMethods: []string{"GET"},
	})
	mux := http.NewServeMux()
	mux.HandleFunc("/", baseHandler)
	mux.HandleFunc("/health_check", healthCheckHandler)
	mux.HandleFunc("/upcoming_streams", getUpcomingStreamsHandler)
	handler := c.Handler(mux)
	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
