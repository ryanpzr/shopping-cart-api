package deleteproduct

import (
	"context"
	"database/sql"

	ordershared "github.com/ryanpzr/shopping-cart-api/internal/order/shared"
	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
	"github.com/ryanpzr/shopping-cart-api/pkg/database"
)

type unitOfWork struct {
	db *sql.DB
}

// NewUnitOfWork binds the product and order repositories to one transaction.
// It lives in the feature because the order module already depends on the
// product repository, so the product package cannot depend on the order one.
func NewUnitOfWork(db *sql.DB) UnitOfWork {
	return &unitOfWork{db: db}
}

func (u *unitOfWork) Do(ctx context.Context, fn func(repos TxRepositories) error) error {
	return database.WithinTx(ctx, u.db, func(tx *sql.Tx) error {
		return fn(TxRepositories{
			Products: shared.NewRepository(tx),
			Orders:   ordershared.NewRepository(tx),
		})
	})
}
