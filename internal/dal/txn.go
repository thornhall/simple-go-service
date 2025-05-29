package dal

import (
	"context"

	"github.com/thornhall/simple-go-service/internal/model"
)

// Run a group of operations inside of a Transaction. If any errors are returned, the transaction is rolled back.
func RunInTx(ctx context.Context, db DB, fn func(ctx context.Context, conn Conn) (*model.UserResponse, error)) (*model.UserResponse, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	res, err := fn(ctx, tx)
	if err != nil {
		return nil, err
	}
	return res, tx.Commit(ctx)
}
