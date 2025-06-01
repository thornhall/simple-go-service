package dal_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/thornhall/simple-go-service/internal/dal"
	"github.com/thornhall/simple-go-service/internal/model"
)

func TestUserRepo_FindById(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()
	assert.NoError(t, err)
	repo := dal.NewUserRepository(mockPool)
	now := time.Now().Truncate(time.Second)

	password, err := bcrypt.GenerateFromPassword([]byte("test_password"), bcrypt.DefaultCost)
	assert.NoError(t, err)
	tests := []struct {
		name      string
		id        int64
		mockSetup func()
		wantUser  *model.User
		wantErr   bool
	}{
		{
			name: "found",
			id:   1234,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{
					"id", "object_id", "first_name", "last_name", "email", "created_at", "updated_at", "password_hash",
				}).AddRow(int64(1234), "uuid-1234", "Alice", "Smith", "a@example.com", now, now, string(password))

				mockPool.
					ExpectQuery(`SELECT id, object_id, first_name, last_name, email, created_at, updated_at, password_hash`).
					WithArgs(int64(1234)).
					WillReturnRows(rows)
			},
			wantUser: &model.User{
				Id:        int64(1234),
				ObjectId:  "uuid-1234",
				FirstName: "Alice",
				LastName:  "Smith",
				Email:     "a@example.com",
				CreatedAt: now,
				UpdatedAt: now,
				Password:  string(password),
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   -1,
			mockSetup: func() {
				mockPool.
					ExpectQuery(`SELECT id, object_id, first_name, last_name, email, created_at, updated_at`).
					WithArgs(int64(-1)).
					WillReturnError(pgx.ErrNoRows)
			},
			wantUser: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			u, err := repo.FindById(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, u)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser, u)
			}
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_FindByObjectID(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()
	assert.NoError(t, err)
	repo := dal.NewUserRepository(mockPool)
	now := time.Now().Truncate(time.Second)

	password, err := bcrypt.GenerateFromPassword([]byte("test_password"), bcrypt.DefaultCost)
	assert.NoError(t, err)
	tests := []struct {
		name      string
		objectID  string
		mockSetup func()
		wantUser  *model.User
		wantErr   bool
	}{
		{
			name:     "found",
			objectID: "uuid-1234",
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{
					"id", "object_id", "first_name", "last_name", "email", "created_at", "updated_at", "password_hash",
				}).AddRow(int64(1), "uuid-1234", "Alice", "Smith", "a@example.com", now, now, string(password))

				mockPool.
					ExpectQuery(`SELECT id, object_id, first_name, last_name, email, created_at, updated_at, password_hash`).
					WithArgs("uuid-1234").
					WillReturnRows(rows)
			},
			wantUser: &model.User{
				Id:        int64(1),
				ObjectId:  "uuid-1234",
				FirstName: "Alice",
				LastName:  "Smith",
				Email:     "a@example.com",
				CreatedAt: now,
				UpdatedAt: now,
				Password:  string(password),
			},
			wantErr: false,
		},
		{
			name:     "not found",
			objectID: "uuid-missing",
			mockSetup: func() {
				mockPool.
					ExpectQuery(`SELECT id, object_id, first_name, last_name, email, created_at, updated_at`).
					WithArgs("uuid-missing").
					WillReturnError(pgx.ErrNoRows)
			},
			wantUser: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			u, err := repo.FindByObjectId(context.Background(), tt.objectID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, u)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser, u)
			}
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_Create(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)

	repo := dal.NewUserRepository(mockPool)
	password, err := bcrypt.GenerateFromPassword([]byte("test_password"), bcrypt.DefaultCost)
	assert.NoError(t, err)
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name      string
		objectId  string
		mockSetup func(objectId string, inputUser *model.User)
		wantErr   bool
	}{
		{
			name:     "success",
			objectId: uuid.New().String(),
			mockSetup: func(objectId string, inputUser *model.User) {
				rows := pgxmock.NewRows([]string{"object_id", "created_at", "updated_at"}).
					AddRow(objectId, now, now)

				mockPool.
					ExpectQuery(`INSERT INTO users.*RETURNING object_id, created_at, updated_at`).
					WithArgs(inputUser.FirstName, inputUser.LastName, inputUser.Email, inputUser.Password).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "query error",
			mockSetup: func(objectId string, inputUser *model.User) {
				mockPool.
					ExpectQuery(`INSERT INTO users.*RETURNING object_id, created_at, updated_at`).
					WithArgs(inputUser.FirstName, inputUser.LastName, inputUser.Email, inputUser.Password).
					WillReturnError(fmt.Errorf("insert failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			inputUser := &model.User{
				FirstName: "Alice",
				LastName:  "Smith",
				Email:     "alice@example.com",
				Password:  string(password),
			}

			testObjectId := uuid.New().String()
			tt.mockSetup(testObjectId, inputUser)

			err := repo.Create(context.Background(), inputUser)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Zero(t, inputUser.ObjectId)
				assert.True(t, inputUser.CreatedAt.IsZero())
				assert.True(t, inputUser.UpdatedAt.IsZero())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testObjectId, inputUser.ObjectId)
				assert.Equal(t, now, inputUser.CreatedAt)
				assert.Equal(t, now, inputUser.UpdatedAt)
			}
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_Update(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)

	repo := dal.NewUserRepository(mockPool)

	baseUser := &model.User{
		Id:        int64(123),
		ObjectId:  "uuid-123",
		FirstName: "Old",
		LastName:  "Name",
		Email:     "old@example.com",
	}

	tests := []struct {
		name      string
		mockSetup func(u *model.User)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "success",
			mockSetup: func(u *model.User) {
				mockPool.
					ExpectExec(`UPDATE users`).
					WithArgs(u.FirstName, u.LastName, u.Email, u.Id).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			wantErr: false,
		},
		{
			name: "no rows affected",
			mockSetup: func(u *model.User) {
				mockPool.
					ExpectExec(`UPDATE users`).
					WithArgs(u.FirstName, u.LastName, u.Email, u.Id).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			wantErr: true,
			errMsg:  fmt.Sprintf("no row updated for object_id=%s", baseUser.ObjectId),
		},
		{
			name: "exec error",
			mockSetup: func(u *model.User) {
				mockPool.
					ExpectExec(`UPDATE users`).
					WithArgs(u.FirstName, u.LastName, u.Email, u.Id).
					WillReturnError(fmt.Errorf("db failure"))
			},
			wantErr: true,
			errMsg:  "db failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := *baseUser
			tt.mockSetup(&u)

			err := repo.Update(context.Background(), &u)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_Delete(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)

	repo := dal.NewUserRepository(mockPool)

	tests := []struct {
		name      string
		id        string
		mockSetup func()
		wantErr   bool
		errMsg    string
	}{
		{
			name: "success",
			id:   "uuid-123",
			mockSetup: func() {
				mockPool.
					ExpectExec(`DELETE FROM users WHERE object_id = \$1`).
					WithArgs("uuid-123").
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			wantErr: false,
		},
		{
			name: "no rows affected",
			id:   "uuid-missing",
			mockSetup: func() {
				mockPool.
					ExpectExec(`DELETE FROM users WHERE object_id = \$1`).
					WithArgs("uuid-missing").
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			wantErr: true,
			errMsg:  "no row deleted for object_id=uuid-missing",
		},
		{
			name: "exec error",
			id:   "uuid-error",
			mockSetup: func() {
				mockPool.
					ExpectExec(`DELETE FROM users WHERE object_id = \$1`).
					WithArgs("uuid-error").
					WillReturnError(fmt.Errorf("db error"))
			},
			wantErr: true,
			errMsg:  "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := repo.Delete(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
