package data

import (
	"context"
	"errors"
	"fmt"

	"PrService/src/internal/application/contracts"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) contracts.TxManager {
	return &TxManager{pool: pool}
}

type contextKey struct{}

var txKey = contextKey{}

func (m *TxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	ctxWithTx := context.WithValue(ctx, txKey, tx)

	if err := fn(ctxWithTx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			return fmt.Errorf("rollback error: %w (original: %w)", rbErr, err)
		}
		return err
	}

	return tx.Commit(ctx)
}

func txFromContext(ctx context.Context) pgx.Tx {
	if v := ctx.Value(txKey); v != nil {
		if tx, ok := v.(pgx.Tx); ok {
			return tx
		}
	}
	return nil
}
