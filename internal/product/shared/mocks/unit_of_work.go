package mocks

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
	outboxmocks "github.com/ryanpzr/shopping-cart-api/pkg/outbox/mocks"
)

// UnitOfWork runs fn with the given mocks and records whether it would have
// committed: a real unit of work commits when fn returns nil.
type UnitOfWork struct {
	Products  *Repository
	Outbox    *outboxmocks.Writer
	Committed bool
}

var _ shared.UnitOfWork = (*UnitOfWork)(nil)

func NewUnitOfWork() *UnitOfWork {
	return &UnitOfWork{Products: new(Repository), Outbox: new(outboxmocks.Writer)}
}

func (u *UnitOfWork) Do(_ context.Context, fn func(repos shared.TxRepositories) error) error {
	err := fn(shared.TxRepositories{Products: u.Products, Outbox: u.Outbox})
	u.Committed = err == nil
	return err
}
