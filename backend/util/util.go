package util

import (
	"log"
	"os"
	"strings"
)

func EnvSetUp() {
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
