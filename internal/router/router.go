package router

import (
	"github.com/gin-gonic/gin"
	"github.com/thornhall/simple-go-service/internal/handler"
	"github.com/thornhall/simple-go-service/internal/service"
)

func RegisterUserRoutes(router *gin.Engine, svc *service.UserService) {
	h := handler.NewUserHandler(svc)
	users := router.Group("/users")
	{
		users.GET("", h.List)
		users.GET("/:object_id", h.Get)
		users.POST("", h.Create)
		users.PUT("/:object_id", h.Update)
		users.DELETE("/:object_id", h.Delete)
	}
}
