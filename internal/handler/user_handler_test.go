package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/thornhall/simple-go-service/internal/dal"
	"github.com/thornhall/simple-go-service/internal/handler"
	"github.com/thornhall/simple-go-service/internal/model"
	"github.com/thornhall/simple-go-service/internal/service"
	"github.com/thornhall/simple-go-service/internal/testutil"
)

func TestMain(m *testing.M) {
	dsn, cleanup := testutil.StartPostgresContainer(&testing.T{})
	defer cleanup()
	os.Setenv("DATABASE_URL", dsn)
	os.Setenv("JWT_SECRET", "test_secret")
	os.Exit(m.Run())
}

func setupRouter(db dal.Conn) *gin.Engine {
	repo := dal.NewUserRepository(db)
	svc := service.NewUserService(repo)
	h := handler.NewUserHandler(svc)

	r := gin.New()
	r.GET("/users/:object_id", h.Get)
	r.POST("/users", h.Create)
	r.PUT("/users/:object_id", h.Update)
	r.DELETE("/users/:object_id", h.Delete)
	return r
}

func TestUserHandler_CRUD(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbURL := os.Getenv("DATABASE_URL")

	maxConns := 25
	maxConnIdleTime := 5 * time.Minute
	db, err := dal.NewPostgresDB(dbURL, maxConns, maxConnIdleTime)
	assert.NoError(t, err)

	router := setupRouter(db)

	createBody := `{"first_name":"Alice","last_name":"Smith","email":"alice@example.com","password":"test_pass"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var created model.UserResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	assert.Equal(t, "Alice", created.FirstName)
	objID := created.ObjectId

	var fn, ln, email string
	err = db.QueryRow(
		context.Background(),
		"SELECT first_name, last_name, email FROM users WHERE object_id = $1",
		objID,
	).Scan(&fn, &ln, &email)
	require.NoError(t, err)
	assert.Equal(t, "Alice", fn)
	assert.Equal(t, "Smith", ln)
	assert.Equal(t, "alice@example.com", email)

	// 3) UPDATE
	updateBody := `{"first_name":"Alicia"}`
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/users/"+objID, bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var updated model.UserResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	assert.Equal(t, "Alicia", updated.FirstName)

	// 4) DELETE
	w = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/users/"+objID, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// 5) GET again → 404
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/users/"+objID, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
