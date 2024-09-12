package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func init() {
	body, err := os.ReadFile(".env")
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

func base_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func main() {
	http.HandleFunc("/", base_handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
