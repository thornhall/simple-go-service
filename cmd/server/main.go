// cmd/server/main.go
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/thornhall/simple-go-service/internal/dal"
	"github.com/thornhall/simple-go-service/internal/router"
	"github.com/thornhall/simple-go-service/internal/service"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to create pgx pool: %v", err)
	}
	defer pool.Close()

	pool.Config().MaxConns = 25
	pool.Config().MaxConnIdleTime = 5 * time.Minute

	userRepo := dal.NewUserRepository(pool)
	userSvc := service.NewUserService(userRepo)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	router.RegisterUserRoutes(r, userSvc)

	log.Println("listening on :8080")
	r.Run(":8080")
}
