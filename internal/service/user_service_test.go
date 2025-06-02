package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/thornhall/simple-go-service/internal/middleware/auth"
	"github.com/thornhall/simple-go-service/internal/model"
)

type fakeRepo struct {
	FindByEmailFunc    func(email string) (*model.User, error)
	FindByIdFunc       func(id int64) (*model.User, error)
	FindByObjectIdFunc func(id string) (*model.User, error)
	CreateFunc         func(u *model.User) error
	UpdateFunc         func(u *model.User) error
	DeleteFunc         func(id string) error
}

func (f *fakeRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return f.FindByEmailFunc(email)
}

func (f *fakeRepo) FindById(ctx context.Context, id int64) (*model.User, error) {
	return f.FindByIdFunc(id)
}
func (f *fakeRepo) FindByObjectId(ctx context.Context, id string) (*model.User, error) {
	return f.FindByObjectIdFunc(id)
}
func (f *fakeRepo) Create(ctx context.Context, u *model.User) error { return f.CreateFunc(u) }
func (f *fakeRepo) Update(ctx context.Context, u *model.User) error { return f.UpdateFunc(u) }
func (f *fakeRepo) Delete(ctx context.Context, id string) error     { return f.DeleteFunc(id) }

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
		FindByObjectIdFunc: func(id string) (*model.User, error) {
			assert.Equal(t, "abc123", id)
			return want, nil
		},
	}
	svc := NewUserService(repo)
	got, err := svc.Get(t.Context(), "abc123")
	require.NoError(t, err)
	assert.Equal(t, want.ObjectId, got.ObjectId)
	assert.Equal(t, want.FirstName, got.FirstName)
	assert.Equal(t, want.LastName, got.LastName)
	assert.Equal(t, want.Email, got.Email)

	// — repo error → ErrNotFound
	repoErr := &fakeRepo{
		FindByObjectIdFunc: func(_ string) (*model.User, error) {
			return nil, errors.New("db is down")
		},
	}
	svc = NewUserService(repoErr)
	_, err = svc.Get(t.Context(), "doesnt-matter")
	assert.Equal(t, ErrNotFound, err)
}

func TestUserService_Login_ReturnsValidJWT(t *testing.T) {
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
	require.NoError(t, err)

	want := &model.User{
		Id:           1,
		ObjectId:     "abc123",
		FirstName:    "Jane",
		LastName:     "Doe",
		Email:        "jane@doe.com",
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		PasswordHash: string(passwordHash),
	}
	repo := &fakeRepo{
		FindByEmailFunc: func(email string) (*model.User, error) {
			assert.Equal(t, "jane@doe.com", email)
			return want, nil
		},
	}

	svc := NewUserService(repo)
	rawToken, err := svc.Login(context.Background(), model.LoginUserInput{
		Email:    "jane@doe.com",
		Password: "test",
	})
	require.NoError(t, err)
	require.NotEmpty(t, rawToken)

	parsed, err := jwt.ParseWithClaims(rawToken, &auth.MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	require.NoError(t, err, "token should parse without error")
	require.NotNil(t, parsed)
	require.True(t, parsed.Valid, "token should be valid")

	// — assert on the claims you expect (e.g. "sub" == user ID, exp is in the future, etc.)
	claims, ok := parsed.Claims.(*auth.MyClaims)
	require.True(t, ok, "claims should be of type MyClaims")
	assert.Equal(t, want.Id, claims.Sub, "`sub` should match the user's ID`")
	assert.False(t, claims.ExpiresAt.Time.Before(time.Now()), "expiration should be in the future")
	assert.False(t, claims.IssuedAt.Time.After(time.Now()), "issued-at should not be in the future")
}

func TestUserService_Create(t *testing.T) {
	var captured *model.User
	repo := &fakeRepo{
		CreateFunc: func(u *model.User) error {
			captured = u
			captured.ObjectId = uuid.NewString()
			return nil
		},
	}
	svc := NewUserService(repo)
	in := model.CreateUserInput{
		FirstName: "Foo",
		LastName:  "Bar",
		Email:     "foo@bar.com",
	}
	resp, jwtSecret, err := svc.Create(t.Context(), in)
	require.NoError(t, err)
	assert.NotNil(t, captured)
	assert.Equal(t, in.FirstName, captured.FirstName)
	assert.Equal(t, in.LastName, captured.LastName)
	assert.Equal(t, in.Email, captured.Email)
	assert.NotNil(t, jwtSecret)

	_, parseErr := uuid.Parse(resp.ObjectId)
	assert.NoError(t, parseErr)
}

func TestUserService_Update(t *testing.T) {
	// — not found
	repoNF := &fakeRepo{
		FindByObjectIdFunc: func(_ string) (*model.User, error) {
			return nil, errors.New("oops")
		},
	}
	svc := NewUserService(repoNF)
	_, err := svc.Update(t.Context(), "id", model.UpdateUserInput{})
	assert.Equal(t, ErrNotFound, err)

	existing := &model.User{
		ObjectId:  "id",
		FirstName: "Orig",
		LastName:  "Name",
		Email:     "orig@x.com",
	}
	var updated *model.User
	repo := &fakeRepo{
		FindByObjectIdFunc: func(_ string) (*model.User, error) {
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
	resp, err := svc.Update(t.Context(), "id", model.UpdateUserInput{
		FirstName: &newFirst,
		Email:     &newEmail,
	})
	require.NoError(t, err)
	assert.Equal(t, "id", resp.ObjectId)
	assert.Equal(t, "NewFirst", resp.FirstName)
	assert.Equal(t, "Name", resp.LastName)
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
	err := svc.Delete(t.Context(), "xyz")
	assert.NoError(t, err)
	assert.Equal(t, "xyz", did)

	// — failure
	repoErr := &fakeRepo{
		DeleteFunc: func(_ string) error {
			return errors.New("cannot delete")
		},
	}
	svc = NewUserService(repoErr)
	err = svc.Delete(t.Context(), "xyz")
	assert.EqualError(t, err, "cannot delete")
}
