package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/thornhall/simple-go-service/internal/middleware/auth"
	"github.com/thornhall/simple-go-service/internal/model"
	"github.com/thornhall/simple-go-service/internal/repo"
)

var ErrNotFound = errors.New("user not found")

type UserService struct {
	repo repo.UserRepository
}

func NewUserService(repo repo.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*model.UserPasswordHash, error) {

	u, err := s.repo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	return &model.UserPasswordHash{PasswordHash: u.Password}, nil
}

func (s *UserService) Get(ctx context.Context, objectId string) (*model.UserResponse, error) {
	u, err := s.repo.FindByObjectId(ctx, objectId)
	if err != nil {
		return nil, ErrNotFound
	}
	return ToUserResponse(u), nil
}

func (s *UserService) Create(ctx context.Context, input model.CreateUserInput) (*model.UserResponse, string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}
	u := &model.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Password:  string(hashed),
	}

	err = s.repo.Create(ctx, u)
	if err != nil {
		return nil, "", err
	}
	signedJwt, err := auth.IssueJWT(u.Id, u.Email)
	if err != nil {
		return nil, "", err
	}
	return ToUserResponse(u), signedJwt, nil
}

func (s *UserService) Update(ctx context.Context, objectId string, input model.UpdateUserInput) (*model.UserResponse, error) {
	u, err := s.repo.FindByObjectId(ctx, objectId)
	if err != nil {
		return nil, ErrNotFound
	}

	if input.FirstName != nil {
		u.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		u.LastName = *input.LastName
	}
	if input.Email != nil {
		u.Email = *input.Email
	}

	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	return ToUserResponse(u), nil
}

func (s *UserService) Delete(ctx context.Context, objectId string) error {
	return s.repo.Delete(ctx, objectId)
}

func ToUserResponse(u *model.User) *model.UserResponse {
	return &model.UserResponse{
		ObjectId:  u.ObjectId,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
	}
}
