package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thornhall/simple-go-service/internal/dal"
	"github.com/thornhall/simple-go-service/internal/router"
	"github.com/thornhall/simple-go-service/internal/service"
)

type Server struct {
	db     dal.DB
	engine *gin.Engine
}

func (s *Server) CloseDB() error {
	s.db.GetPool().Close()
	return nil
}

func NewServer(dbURL string, jwtSecretStr string) (*Server, error) {
	maxConns := 25
	maxConnIdleTime := 5 * time.Minute
	db, err := dal.NewPostgresDB(dbURL, maxConns, maxConnIdleTime)
	if err != nil {
		return nil, err
	}
	repo := dal.NewUserRepository(db)
	userSvc := service.NewUserService(repo)

	r := gin.New()

	//authMiddleware := r.Group("/")
	//authMiddleware.Use(auth.JWTAuth(userSvc, []byte(jwtSecretStr)))

	r.Use(gin.Logger(), gin.Recovery())
	router.RegisterUserRoutes(r, userSvc)

	server := &Server{
		db:     db,
		engine: r,
	}
	return server, nil
}
