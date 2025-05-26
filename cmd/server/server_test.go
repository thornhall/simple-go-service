package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewServer_WithRealDB(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	gin.SetMode(gin.TestMode)
	server, err := NewServer(dbURL)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users", nil)
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, "[]", w.Body.String())
}
