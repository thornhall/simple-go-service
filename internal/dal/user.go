package dal

import (
	"context"
	"fmt"

	"github.com/thornhall/simple-go-service/internal/model"
	"github.com/thornhall/simple-go-service/internal/service"
)

type UserRepo struct {
	db DB
}

func NewUserRepository(db DB) service.UserRepository {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByID(objectId string) (*model.User, error) {
	const sql = `
SELECT id, object_id, first_name, last_name, email, created_at, updated_at
  FROM users
WHERE object_id = $1
`
	u := &model.User{}
	err := r.db.QueryRow(context.Background(), sql, objectId).
		Scan(&u.Id, &u.ObjectId, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// FindAll returns all users ordered by creation time (newest first).
func (r *UserRepo) FindAll() ([]*model.User, error) {
	const sql = `
SELECT id, object_id, first_name, last_name, email, created_at, updated_at
  FROM users
ORDER BY created_at DESC
`

	rows, err := r.db.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		if err := rows.Scan(&u.Id, &u.ObjectId, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepo) Create(u *model.User) error {
	const sql = `
INSERT INTO users (object_id, first_name, last_name, email)
VALUES (get_random_uuid(), $1, $2, $3)
RETURNING object_id, created_at, updated_at
`
	row := r.db.QueryRow(context.Background(), sql,
		u.FirstName, u.LastName, u.Email,
	)
	return row.Scan(&u.ObjectId, &u.CreatedAt, &u.UpdatedAt)
}

// Update applies changes to an existing user identified by ObjectId.
func (r *UserRepo) Update(u *model.User) error {
	const sql = `
UPDATE users
   SET first_name = $1,
       last_name  = $2,
       email      = $3,
       updated_at = now()
 WHERE id = $4
`
	cmd, err := r.db.Exec(context.Background(), sql,
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

func (r *UserRepo) Delete(id string) error {
	const sql = `DELETE FROM users WHERE object_id = $1`
	cmd, err := r.db.Exec(context.Background(), sql, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("no row deleted for object_id=%s", id)
	}
	return nil
}
