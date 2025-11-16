package data

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxQuerier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func QuerierFromContext(ctx context.Context, pool *pgxpool.Pool) PgxQuerier {
	if tx := txFromContext(ctx); tx != nil {
		return tx
	}
	return pool
}

const (
	pgCodeUniqueViolation = "23505"
)

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgCodeUniqueViolation
	}
	return false
}

func IsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
