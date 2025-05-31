package router_test

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/thornhall/simple-go-service/internal/dal"
	"github.com/thornhall/simple-go-service/internal/router"
	"github.com/thornhall/simple-go-service/internal/service"
)

func TestRegisterUserRoutes_RegistersAllEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	var noopDB dal.Conn
	repo := dal.NewUserRepository(noopDB)
	svc := service.NewUserService(repo)
	router.RegisterUserRoutes(r, svc)

	routes := r.Routes()
	expected := []struct {
		method, path string
	}{
		{"GET", "/users/:object_id"},
		{"POST", "/users"},
		{"PUT", "/users/:object_id"},
		{"DELETE", "/users/:object_id"},
	}

	for _, exp := range expected {
		found := false
		for _, rt := range routes {
			if rt.Method == exp.method && rt.Path == exp.path {
				found = true
				break
			}
		}
		assert.Truef(t, found, "%s %s not registered", exp.method, exp.path)
	}
}
