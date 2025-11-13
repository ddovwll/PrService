package data

import (
	"context"
	"database/sql"
	"fmt"

	"PrService/src/internal/application"

	"github.com/jmoiron/sqlx"
)

type txManager struct {
	db *sqlx.DB
}

func NewSQLXManager(db *sqlx.DB) application.TxManager {
	return &txManager{db: db}
}

type contextKey struct{}

var txKey = contextKey{}

func (m *txManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	ctxWithTx := context.WithValue(ctx, txKey, tx)

	if err := fn(ctxWithTx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback error: %w (original: %w)", rbErr, err)
		}
		return err
	}

	return tx.Commit()
}
