package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/thornhall/simple-go-service/internal/dal"
	"github.com/thornhall/simple-go-service/internal/router"
	"github.com/thornhall/simple-go-service/internal/service"
)

func NewServer(dbURL string) (*gin.Engine, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.Connect(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	pool.Config().MaxConns = 25
	pool.Config().MaxConnIdleTime = 5 * time.Minute

	userRepo := dal.NewUserRepository(pool)
	userSvc := service.NewUserService(userRepo)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	router.RegisterUserRoutes(r, userSvc)

	return r, nil
}
