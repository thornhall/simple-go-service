package model

import (
	"time"
)

type User struct {
	Id        int       `json:"id"`
	ObjectId  string    `json:"object_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`
	Email     string    `json:"email"`
}

// POST
type CreateUserInputDTO struct {
	FirstName string `json:"first_name"  binding:"required"`
	LastName  string `json:"last_name"`
	Email     string `json:"email" binding:"required,email"`
}

// PUT
type UpdateUserInputDTO struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Email     *string `json:"email,omitempty"`
}
