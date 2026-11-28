package shared

import (
	"context"
	"database/sql"

	"github.com/ryanpzr/shopping-cart-api/pkg/database"
	"github.com/ryanpzr/shopping-cart-api/pkg/outbox"
)

// TxRepositories are repositories bound to the same transaction. The outbox
// writer is bound to it too, so an audit event is committed with the change it
// describes, or not at all.
type TxRepositories struct {
	Products Repository
	Outbox   outbox.Writer
}

// UnitOfWork runs fn inside one transaction: it commits when fn returns nil
// and rolls back otherwise. Usecases depend on this port, so they stay
// testable without a database.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(repos TxRepositories) error) error
}

type unitOfWork struct {
	db *sql.DB
}

func NewUnitOfWork(db *sql.DB) UnitOfWork {
	return &unitOfWork{db: db}
}

func (u *unitOfWork) Do(ctx context.Context, fn func(repos TxRepositories) error) error {
	return database.WithinTx(ctx, u.db, func(tx *sql.Tx) error {
		return fn(TxRepositories{
			Products: NewRepository(tx),
			Outbox:   outbox.NewWriter(tx),
		})
	})
}
