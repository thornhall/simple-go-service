package dal

import (
	"context"
	"fmt"

	"github.com/thornhall/simple-go-service/internal/model"
	"github.com/thornhall/simple-go-service/internal/repo"
)

type UserRepo struct {
	conn Conn
}

func NewUserRepository(conn Conn) repo.UserRepository {
	return &UserRepo{conn: conn}
}

func (r *UserRepo) FindByID(ctx context.Context, objectId string) (*model.User, error) {
	const sql = `
SELECT id, object_id, first_name, last_name, email, created_at, updated_at, password_hash
  FROM users
WHERE object_id = $1
`
	u := &model.User{}
	err := r.conn.QueryRow(context.Background(), sql, objectId).
		Scan(&u.Id, &u.ObjectId, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt, &u.UpdatedAt, &u.Password)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	const sql = `
INSERT INTO users (first_name, last_name, email, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING object_id, created_at, updated_at
`
	row := r.conn.QueryRow(context.Background(), sql,
		u.FirstName, u.LastName, u.Email, u.Password,
	)
	return row.Scan(&u.ObjectId, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepo) Update(ctx context.Context, u *model.User) error {
	const sql = `
UPDATE users
   SET first_name = $1,
       last_name  = $2,
       email      = $3,
       updated_at = now()
 WHERE id = $4
`
	cmd, err := r.conn.Exec(context.Background(), sql,
		u.FirstName, u.LastName, u.Email, u.Id,
	)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("no row updated for object_id=%s", u.ObjectId)
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, objectId string) error {
	const sql = `DELETE FROM users WHERE object_id = $1`
	cmd, err := r.conn.Exec(context.Background(), sql, objectId)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("no row deleted for object_id=%s", objectId)
	}
	return nil
}
