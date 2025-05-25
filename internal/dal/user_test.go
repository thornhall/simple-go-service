package dal_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"

	"github.com/thornhall/simple-go-service/internal/dal"
	"github.com/thornhall/simple-go-service/internal/model"
)

func TestUserRepo_FindByID(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := dal.NewUserRepository(mockPool)
	now := time.Now().Truncate(time.Second)

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
					"id", "object_id", "first_name", "last_name", "email", "created_at", "updated_at",
				}).AddRow(1, "uuid-1234", "Alice", "Smith", "a@example.com", now, now)

				mockPool.
					ExpectQuery(`SELECT id, object_id, first_name, last_name, email, created_at, updated_at`).
					WithArgs("uuid-1234").
					WillReturnRows(rows)
			},
			wantUser: &model.User{
				Id:        1,
				ObjectId:  "uuid-1234",
				FirstName: "Alice",
				LastName:  "Smith",
				Email:     "a@example.com",
				CreatedAt: now,
				UpdatedAt: now,
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

			u, err := repo.FindByID(tt.objectID)
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

func TestUserRepo_FindAll(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := dal.NewUserRepository(mockPool)

	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name      string
		mockSetup func()
		wantUsers []*model.User
		wantErr   bool
	}{
		{
			name: "multiple rows",
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{
					"id", "object_id", "first_name", "last_name",
					"email", "created_at", "updated_at",
				})
				for i := 1; i <= 3; i++ {
					rows.AddRow(
						i,
						"uuid-"+strconv.Itoa(i),
						"Alice"+strconv.Itoa(i),
						"Smith",
						"a"+strconv.Itoa(i)+"@example.com",
						now,
						now,
					)
				}
				mockPool.
					ExpectQuery(`SELECT id, object_id, first_name, last_name, email, created_at, updated_at`).
					WillReturnRows(rows)
			},
			wantUsers: []*model.User{
				{Id: 1, ObjectId: "uuid-1", FirstName: "Alice1", LastName: "Smith", Email: "a1@example.com", CreatedAt: now, UpdatedAt: now},
				{Id: 2, ObjectId: "uuid-2", FirstName: "Alice2", LastName: "Smith", Email: "a2@example.com", CreatedAt: now, UpdatedAt: now},
				{Id: 3, ObjectId: "uuid-3", FirstName: "Alice3", LastName: "Smith", Email: "a3@example.com", CreatedAt: now, UpdatedAt: now},
			},
			wantErr: false,
		},
		{
			name: "empty result",
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{
					"id", "object_id", "first_name", "last_name",
					"email", "created_at", "updated_at",
				})
				mockPool.
					ExpectQuery(`SELECT id, object_id, first_name, last_name, email, created_at, updated_at`).
					WillReturnRows(rows)
			},
			wantUsers: []*model.User(nil),
			wantErr:   false,
		},
		{
			name: "query error",
			mockSetup: func() {
				mockPool.
					ExpectQuery(`SELECT id, object_id, first_name, last_name, email, created_at, updated_at`).
					WillReturnError(fmt.Errorf("db is down"))
			},
			wantUsers: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			users, err := repo.FindAll()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, users)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUsers, users)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
