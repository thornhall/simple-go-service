package main

import (
	"log"
	"os"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	engine, err := NewServer(dbURL)
	if err != nil {
		log.Fatalf("failed to build server: %v", err)
	}

	log.Println("listening on :8080")
	engine.Run(":8080")
}
