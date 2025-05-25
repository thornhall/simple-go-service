package service

import (
	"errors"
	"github.com/google/uuid"
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

func (s *UserService) Get(objectId string) (*model.User, error) {
	u, err := s.repo.FindByID(objectId)
	if err != nil {
		return nil, ErrNotFound
	}
	return u, nil
}

func (s *UserService) List() ([]*model.User, error) {
	return s.repo.FindAll()
}

func (s *UserService) Create(input model.CreateUserInputDTO) (*model.User, error) {
	if input.FirstName == "" || input.Email == "" {
		return nil, errors.New("name and email are required")
	}

	u := &model.User{
		ObjectId:  uuid.New().String(),
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
	}

	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) Update(objectId string, input model.UpdateUserInputDTO) (*model.User, error) {
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
	return u, nil
}

func (s *UserService) Delete(objectId string) error {
	return s.repo.Delete(objectId)
}
