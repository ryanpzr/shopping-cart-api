// Package mocks provides test doubles for the cart repositories and the unit
// of work, shared by the cart feature tests.
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/domain"
	"github.com/ryanpzr/shopping-cart-api/internal/cart/shared"
	ordermocks "github.com/ryanpzr/shopping-cart-api/internal/order/shared/mocks"
	productmocks "github.com/ryanpzr/shopping-cart-api/internal/product/shared/mocks"
	outboxmocks "github.com/ryanpzr/shopping-cart-api/pkg/outbox/mocks"
)

type Repository struct{ mock.Mock }

var _ shared.Repository = (*Repository)(nil)

func (m *Repository) FindByBuyer(ctx context.Context, buyerID int, status string) (domain.Cart, error) {
	args := m.Called(ctx, buyerID, status)
	cart, _ := args.Get(0).(domain.Cart)
	return cart, args.Error(1)
}

func (m *Repository) LockByBuyer(ctx context.Context, buyerID int, status string) (domain.Cart, error) {
	args := m.Called(ctx, buyerID, status)
	cart, _ := args.Get(0).(domain.Cart)
	return cart, args.Error(1)
}

func (m *Repository) Create(ctx context.Context, buyerID int, status string) (domain.Cart, error) {
	args := m.Called(ctx, buyerID, status)
	cart, _ := args.Get(0).(domain.Cart)
	return cart, args.Error(1)
}

func (m *Repository) UpdateStatus(ctx context.Context, cartID int, status string) error {
	return m.Called(ctx, cartID, status).Error(0)
}

func (m *Repository) AddItem(ctx context.Context, item domain.Item) (domain.Item, error) {
	args := m.Called(ctx, item)
	added, _ := args.Get(0).(domain.Item)
	return added, args.Error(1)
}

func (m *Repository) UpdateItem(ctx context.Context, item domain.Item) error {
	return m.Called(ctx, item).Error(0)
}

func (m *Repository) RemoveItem(ctx context.Context, cartID, itemID int) error {
	return m.Called(ctx, cartID, itemID).Error(0)
}

// UnitOfWork runs fn with the given mocks and records whether it would have
// committed: a real unit of work commits when fn returns nil.
type UnitOfWork struct {
	Carts     *Repository
	Products  *productmocks.Repository
	Orders    *ordermocks.Repository
	Outbox    *outboxmocks.Writer
	Committed bool
}

func NewUnitOfWork() *UnitOfWork {
	return &UnitOfWork{
		Carts:    new(Repository),
		Products: new(productmocks.Repository),
		Orders:   new(ordermocks.Repository),
		Outbox:   new(outboxmocks.Writer),
	}
}

func (u *UnitOfWork) Do(_ context.Context, fn func(repos shared.TxRepositories) error) error {
	err := fn(shared.TxRepositories{Carts: u.Carts, Products: u.Products, Orders: u.Orders, Outbox: u.Outbox})
	u.Committed = err == nil
	return err
}
