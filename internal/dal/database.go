package dal

import (
	"context"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Conn interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

type Tx interface {
	Conn
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type DB interface {
	Conn
	Begin(ctx context.Context) (Tx, error)
	GetPool() *pgxpool.Pool
}

type pgxDB struct {
	*pgxpool.Pool
}

func NewPostgresDB(connString string, maxConns int, maxConnIdleTime time.Duration) (DB, error) {
	pool, err := pgxpool.Connect(context.Background(), connString)
	pool.Config().MaxConnIdleTime = maxConnIdleTime
	pool.Config().MaxConns = int32(maxConns)
	if err != nil {
		return nil, err
	}
	return &pgxDB{pool}, nil
}

func (p *pgxDB) Begin(ctx context.Context) (Tx, error) {
	return p.Pool.Begin(ctx)
}

func (p *pgxDB) Config() *pgxpool.Config {
	return p.Pool.Config()
}

func (p *pgxDB) GetPool() *pgxpool.Pool {
	return p.Pool
}
