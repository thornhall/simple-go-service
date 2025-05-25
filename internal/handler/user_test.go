package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/thornhall/simple-go-service/internal/handler"
	"github.com/thornhall/simple-go-service/internal/model"
	"github.com/thornhall/simple-go-service/internal/service"
)

type InMemoryRepo struct {
	users map[string]*model.User
}

var _ service.UserRepository = (*InMemoryRepo)(nil)

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{users: make(map[string]*model.User)}
}

func (r *InMemoryRepo) FindByID(id string) (*model.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, service.ErrNotFound
	}
	return u, nil
}

func (r *InMemoryRepo) FindAll() ([]*model.User, error) {
	list := make([]*model.User, 0, len(r.users))
	for _, u := range r.users {
		list = append(list, u)
	}
	return list, nil
}

func (r *InMemoryRepo) Create(u *model.User) error {
	u.Id = len(r.users) + 1
	u.ObjectId = fmt.Sprintf("uuid-%d", u.Id)
	now := time.Now().UTC().Truncate(time.Second)
	u.CreatedAt = now
	u.UpdatedAt = now
	copy := *u
	r.users[u.ObjectId] = &copy
	return nil
}

func (r *InMemoryRepo) Update(u *model.User) error {
	existing, ok := r.users[u.ObjectId]
	if !ok {
		return service.ErrNotFound
	}
	existing.FirstName = u.FirstName
	existing.LastName = u.LastName
	existing.Email = u.Email
	existing.UpdatedAt = time.Now().UTC().Truncate(time.Second)
	return nil
}

func (r *InMemoryRepo) Delete(id string) error {
	if _, ok := r.users[id]; !ok {
		return service.ErrNotFound
	}
	delete(r.users, id)
	return nil
}

func setupRouter() *gin.Engine {
	repo := NewInMemoryRepo()
	svc := service.NewUserService(repo)
	h := handler.NewUserHandler(svc)

	r := gin.New()
	r.GET("/users", h.List)
	r.GET("/users/:object_id", h.Get)
	r.POST("/users", h.Create)
	r.PUT("/users/:object_id", h.Update)
	r.DELETE("/users/:object_id", h.Delete)
	return r
}

func TestUserHandler_CRUD(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, "[]", w.Body.String())

	createBody := `{"first_name":"Alice","last_name":"Smith","email":"alice@example.com"}`
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/users", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var created model.UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &created)
	assert.NoError(t, err)
	assert.Equal(t, "Alice", created.FirstName)
	assert.Equal(t, "Smith", created.LastName)
	assert.Equal(t, "alice@example.com", created.Email)
	objID := created.ObjectId

	// 3) GET that user back
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/users/"+objID, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var fetched model.UserResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &fetched))
	assert.Equal(t, created, fetched)

	// 4) UPDATE the user
	updateBody := `{"first_name":"Alicia"}`
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/users/"+objID, bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var updated model.User
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	assert.Equal(t, "Alicia", updated.FirstName)
	assert.Equal(t, "Smith", updated.LastName)

	// 5) DELETE the user
	w = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/users/"+objID, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// 6) GET again → 404
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/users/"+objID, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
