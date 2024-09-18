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

func getUpcomingStreamsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	origin := r.Header.Get("Origin")
	fmt.Printf("Request originated from: %s\n", origin)
	results := api.GetAllUpcomingStreams()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func main() {
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"https://*.onrender.com"},
		AllowedMethods: []string{"GET"},
	})
	mux := http.NewServeMux()
	mux.HandleFunc("/", baseHandler)
	mux.HandleFunc("/upcoming_streams", getUpcomingStreamsHandler)
	handler := c.Handler(mux)
	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
