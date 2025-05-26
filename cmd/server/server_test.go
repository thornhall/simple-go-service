package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewServer_WithRealDB(t *testing.T) {
	dbURL := "postgres://user:pass@localhost:5433/simple_service?sslmode=disable"

	if err := RunMigrations(dbURL); err != nil {
		t.Fatalf("migrations failed: %v", err)
	}

	gin.SetMode(gin.TestMode)
	server, err := NewServer(dbURL)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users", nil)
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, "[]", w.Body.String())
}
