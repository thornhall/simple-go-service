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
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	server, err := NewServer(dbURL, jwtSecret)
	if err != nil {
		log.Fatalf("failed to build server: %v", err)
	}

	log.Println("listening on :8080")
	server.engine.Run(":8080")
}
