package repo

import (
	"context"
	"github.com/thornhall/simple-go-service/internal/model"
)

type UserRepository interface {
	FindById(ctx context.Context, userId int64) (*model.User, error)
	FindByObjectId(ctx context.Context, objectID string) (*model.User, error)
	Create(ctx context.Context, u *model.User) error
	Update(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, objectID string) error
}
