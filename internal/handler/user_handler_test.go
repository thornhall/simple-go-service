package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/thornhall/simple-go-service/internal/dal"
	"github.com/thornhall/simple-go-service/internal/handler"
	"github.com/thornhall/simple-go-service/internal/model"
	"github.com/thornhall/simple-go-service/internal/service"
)

func runMigrations(t *testing.T, sqlDB *sql.DB) {
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	require.NoError(t, err)

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../db/migrations", // adjust this path
		"postgres",
		driver,
	)
	require.NoError(t, err)

	// Apply all up migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Up(): %v", err)
	}
}

func setupTestPostgres(t *testing.T) dal.DB {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image: "postgres:15-alpine",
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "secret",
			"POSTGRES_DB":       "testdb",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	pgC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { pgC.Terminate(ctx) })

	host, _ := pgC.Host(ctx)
	port, _ := pgC.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf(
		"postgres://postgres:secret@%s:%s/testdb?sslmode=disable",
		host, port.Port(),
	)

	// 1) Open a *sql.DB* via pgx
	sqlDB, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	require.NoError(t, sqlDB.Ping())

	// 2) Run file-based migrations
	runMigrations(t, sqlDB)

	db, err := dal.NewPostgresDB(dsn, 5, time.Minute)
	require.NoError(t, err)
	return db
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

	db := setupTestPostgres(t)
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

	// 2) VERIFY in the database
	var fn, ln, email string
	err := db.QueryRow(
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
