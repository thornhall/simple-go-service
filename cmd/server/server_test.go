package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/thornhall/simple-go-service/internal/model"
	"github.com/thornhall/simple-go-service/internal/testutil"
)

func TestMain(m *testing.M) {
	dsn, container := testutil.StartPostgresContainer(&testing.T{})
	os.Setenv("DATABASE_URL", dsn)
	os.Setenv("JWT_SECRET", "test_secret")
	code := m.Run()
	container.Terminate(context.Background())
	os.Exit(code)
}

func TestNewServer_WithRealDB(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	assert.NotZero(t, dbURL)
	assert.NotZero(t, jwtSecret)
	gin.SetMode(gin.TestMode)
	server, err := NewServer(dbURL, jwtSecret)
	assert.NoError(t, err)

	createBody := `{"first_name":"S","last_name":"Smith","email":"alices@example.com","password":"test_pass"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users", bytes.NewBufferString(createBody))
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var userResponse model.UserResponse
	err = json.Unmarshal(w.Body.Bytes(), &userResponse)
	assert.NoError(t, err)
	assert.Equal(t, "S", userResponse.FirstName)
	assert.Equal(t, "alices@example.com", userResponse.Email)
	assert.Equal(t, "Smith", userResponse.LastName)
	_, err = uuid.Parse(userResponse.ObjectId)
	assert.NoError(t, err)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/users/"+userResponse.ObjectId, nil)
	server.ServeHTTP(w, req)
}
