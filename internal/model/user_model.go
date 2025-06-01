package model

import (
	"time"
)

type User struct {
	Id        int64     `db:"id"`
	ObjectId  string    `db:"object_id"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	IsDeleted bool      `db:"is_deleted"`
	Email     string    `db:"email"`
	Password  string    `db:"password_hash"`
}

type UserCreateResponse struct {
	ObjectId     string `json:"object_id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Email        string `json:"email"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type UserResponse struct {
	ObjectId  string `json:"object_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type CreateUserResponse struct {
	*UserResponse
	JWT string `json:"jwt"`
}

type UserPasswordHash struct {
	PasswordHash string `json:"password_hash"`
}

// POST
type CreateUserInput struct {
	FirstName string `json:"first_name"  binding:"required"`
	LastName  string `json:"last_name"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8,max=64"`
}

// PUT
type UpdateUserInput struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Email     *string `json:"email,omitempty"`
}
