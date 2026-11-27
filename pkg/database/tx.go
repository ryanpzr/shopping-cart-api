// Package database holds helpers shared by the repositories.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// DBTX is satisfied by both *sql.DB and *sql.Tx, so repositories can run
// inside a transaction owned by the caller.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// WithinTx runs fn inside a transaction. It commits when fn returns nil and
// rolls back when fn returns an error or panics.
func WithinTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			return errors.Join(err, fmt.Errorf("rollback: %w", rbErr))
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
