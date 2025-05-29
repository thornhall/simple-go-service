package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	//"github.com/thornhall/simple-go-service/internal/middleware/auth"
	"github.com/thornhall/simple-go-service/internal/model"
	"github.com/thornhall/simple-go-service/internal/service"
)

type UserHandler struct {
	Svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{Svc: svc}
}

func (h *UserHandler) Get(ctx *gin.Context) {
	objectId := ctx.Param("object_id")
	user, err := h.Svc.Get(ctx, objectId)
	if err == service.ErrNotFound {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (h *UserHandler) Create(ctx *gin.Context) {
	var input model.CreateUserInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		if errors.Is(err, io.EOF) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "request body cannot be empty"})
			return
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	user, err := h.Svc.Create(ctx, input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//token, err := auth.GenerateToken(user.ObjectId)
	// if err != nil {
	// 	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not create token"})
	// 	return
	// }

	ctx.JSON(http.StatusCreated, user)
}

func (h *UserHandler) Update(ctx *gin.Context) {
	objectId := ctx.Param("object_id")
	var input model.UpdateUserInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.Svc.Update(ctx, objectId, input)
	if err == service.ErrNotFound {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (h *UserHandler) Delete(ctx *gin.Context) {
	objectId := ctx.Param("object_id")
	if err := h.Svc.Delete(ctx, objectId); err == service.ErrNotFound {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		ctx.Status(http.StatusNoContent)
	}
}
