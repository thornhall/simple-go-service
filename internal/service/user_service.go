package service

import (
	"errors"
	"github.com/thornhall/simple-go-service/internal/model"
)

var ErrNotFound = errors.New("user not found")

type UserRepository interface {
	FindByID(id string) (*model.User, error)
	FindAll() ([]*model.User, error)
	Create(u *model.User) error
	Update(u *model.User) error
	Delete(objectId string) error
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Get(objectId string) (*model.UserResponse, error) {
	u, err := s.repo.FindByID(objectId)
	if err != nil {
		return nil, ErrNotFound
	}
	return ToUserResponse(u), nil
}

func (s *UserService) List() ([]*model.UserResponse, error) {
	users, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	userResponse := []*model.UserResponse{}
	for _, user := range users {
		userResponse = append(userResponse, ToUserResponse(user))
	}
	return userResponse, nil
}

func (s *UserService) Create(input model.CreateUserInput) (*model.UserResponse, error) {
	if input.FirstName == "" || input.Email == "" {
		return nil, errors.New("name and email are required")
	}

	u := &model.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
	}

	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	return ToUserResponse(u), nil
}

func (s *UserService) Update(objectId string, input model.UpdateUserInput) (*model.UserResponse, error) {
	u, err := s.repo.FindByID(objectId)
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

	if err := s.repo.Update(u); err != nil {
		return nil, err
	}
	return ToUserResponse(u), nil
}

func (s *UserService) Delete(objectId string) error {
	return s.repo.Delete(objectId)
}

func ToUserResponse(u *model.User) *model.UserResponse {
	return &model.UserResponse{
		ObjectId:  u.ObjectId,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
