// internal/service/user_service_test.go
package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/thornhall/simple-go-service/internal/model"
)

type fakeRepo struct {
	FindByIDFunc func(id string) (*model.User, error)
	FindAllFunc  func() ([]*model.User, error)
	CreateFunc   func(u *model.User) error
	UpdateFunc   func(u *model.User) error
	DeleteFunc   func(id string) error
}

func (f *fakeRepo) FindByID(id string) (*model.User, error) { return f.FindByIDFunc(id) }
func (f *fakeRepo) FindAll() ([]*model.User, error)         { return f.FindAllFunc() }
func (f *fakeRepo) Create(u *model.User) error              { return f.CreateFunc(u) }
func (f *fakeRepo) Update(u *model.User) error              { return f.UpdateFunc(u) }
func (f *fakeRepo) Delete(id string) error                  { return f.DeleteFunc(id) }

func TestUserService_Get(t *testing.T) {
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()

	// — success case
	want := &model.User{
		ObjectId:  "abc123",
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@doe.com",
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	repo := &fakeRepo{
		FindByIDFunc: func(id string) (*model.User, error) {
			assert.Equal(t, "abc123", id)
			return want, nil
		},
	}
	svc := NewUserService(repo)
	got, err := svc.Get("abc123")
	assert.NoError(t, err)
	assert.Equal(t, want.ObjectId, got.ObjectId)
	assert.Equal(t, want.FirstName, got.FirstName)
	assert.Equal(t, want.LastName, got.LastName)
	assert.Equal(t, want.Email, got.Email)
	assert.Equal(t, want.CreatedAt, got.CreatedAt)
	assert.Equal(t, want.UpdatedAt, got.UpdatedAt)

	// — repo error → ErrNotFound
	repoErr := &fakeRepo{
		FindByIDFunc: func(_ string) (*model.User, error) {
			return nil, errors.New("db is down")
		},
	}
	svc = NewUserService(repoErr)
	_, err = svc.Get("doesnt-matter")
	assert.Equal(t, ErrNotFound, err)
}

func TestUserService_List(t *testing.T) {
	now := time.Now()

	// — error path
	repoErr := &fakeRepo{
		FindAllFunc: func() ([]*model.User, error) {
			return nil, errors.New("cannot list")
		},
	}
	svc := NewUserService(repoErr)
	_, err := svc.List()
	assert.EqualError(t, err, "cannot list")

	// — success
	users := []*model.User{
		{ObjectId: uuid.NewString(), FirstName: "A", Email: "a@", CreatedAt: now, UpdatedAt: now},
		{ObjectId: uuid.NewString(), FirstName: "B", Email: "b@", CreatedAt: now, UpdatedAt: now},
	}
	repoOK := &fakeRepo{
		FindAllFunc: func() ([]*model.User, error) {
			return users, nil
		},
	}
	svc = NewUserService(repoOK)
	resps, err := svc.List()
	assert.NoError(t, err)
	assert.Len(t, resps, 2)

	for i, u := range users {
		assert.Equal(t, u.ObjectId, resps[i].ObjectId)
		assert.Equal(t, u.FirstName, resps[i].FirstName)
		assert.Equal(t, u.Email, resps[i].Email)
		assert.Equal(t, u.CreatedAt, resps[i].CreatedAt)
		assert.Equal(t, u.UpdatedAt, resps[i].UpdatedAt)
	}
}

func TestUserService_Create(t *testing.T) {
	svc := NewUserService(&fakeRepo{})
	_, err := svc.Create(model.CreateUserInput{FirstName: "", Email: ""})
	assert.EqualError(t, err, "name and email are required")

	_, err = svc.Create(model.CreateUserInput{FirstName: "X", Email: ""})
	assert.EqualError(t, err, "name and email are required")

	var captured *model.User
	repo := &fakeRepo{
		CreateFunc: func(u *model.User) error {
			captured = u
			captured.ObjectId = uuid.NewString()
			return nil
		},
	}
	svc = NewUserService(repo)
	in := model.CreateUserInput{
		FirstName: "Foo",
		LastName:  "Bar",
		Email:     "foo@bar.com",
	}
	resp, err := svc.Create(in)
	assert.NoError(t, err)
	assert.NotNil(t, captured)
	assert.Equal(t, in.FirstName, captured.FirstName)
	assert.Equal(t, in.LastName, captured.LastName)
	assert.Equal(t, in.Email, captured.Email)

	_, parseErr := uuid.Parse(resp.ObjectId)
	assert.NoError(t, parseErr)
}

func TestUserService_Update(t *testing.T) {
	// — not found
	repoNF := &fakeRepo{
		FindByIDFunc: func(_ string) (*model.User, error) {
			return nil, errors.New("oops")
		},
	}
	svc := NewUserService(repoNF)
	_, err := svc.Update("id", model.UpdateUserInput{})
	assert.Equal(t, ErrNotFound, err)

	// — partial update
	existing := &model.User{
		ObjectId:  "id",
		FirstName: "Orig",
		LastName:  "Name",
		Email:     "orig@x.com",
	}
	var updated *model.User
	repo := &fakeRepo{
		FindByIDFunc: func(_ string) (*model.User, error) {
			return existing, nil
		},
		UpdateFunc: func(u *model.User) error {
			updated = u
			return nil
		},
	}
	svc = NewUserService(repo)
	newFirst := "NewFirst"
	newEmail := "new@x.com"
	resp, err := svc.Update("id", model.UpdateUserInput{
		FirstName: &newFirst,
		Email:     &newEmail,
	})
	assert.NoError(t, err)

	assert.Equal(t, "id", resp.ObjectId)
	assert.Equal(t, "NewFirst", resp.FirstName)
	assert.Equal(t, "Name", resp.LastName) // unchanged
	assert.Equal(t, "new@x.com", resp.Email)
	assert.Len(t, []string{updated.FirstName, updated.Email}, 2)
}

func TestUserService_Delete(t *testing.T) {
	// — success
	var did string
	repoOK := &fakeRepo{
		DeleteFunc: func(id string) error {
			did = id
			return nil
		},
	}
	svc := NewUserService(repoOK)
	err := svc.Delete("xyz")
	assert.NoError(t, err)
	assert.Equal(t, "xyz", did)

	// — failure
	repoErr := &fakeRepo{
		DeleteFunc: func(_ string) error {
			return errors.New("cannot delete")
		},
	}
	svc = NewUserService(repoErr)
	err = svc.Delete("xyz")
	assert.EqualError(t, err, "cannot delete")
}
