package shared

import (
	"context"
	"database/sql"

	ordershared "github.com/ryanpzr/shopping-cart-api/internal/order/shared"
	productshared "github.com/ryanpzr/shopping-cart-api/internal/product/shared"
	"github.com/ryanpzr/shopping-cart-api/pkg/database"
)

// TxRepositories are repositories bound to the same transaction.
type TxRepositories struct {
	Carts    Repository
	Products productshared.Repository
	Orders   ordershared.Repository
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
			Carts:    NewRepository(tx),
			Products: productshared.NewRepository(tx),
			Orders:   ordershared.NewRepository(tx),
		})
	})
}
