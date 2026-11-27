// Package mocks provides test doubles for the cart repositories and the unit
// of work, shared by the cart feature tests.
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/domain"
	"github.com/ryanpzr/shopping-cart-api/internal/cart/shared"
	orderdomain "github.com/ryanpzr/shopping-cart-api/internal/order/domain"
	ordershared "github.com/ryanpzr/shopping-cart-api/internal/order/shared"
	productmocks "github.com/ryanpzr/shopping-cart-api/internal/product/shared/mocks"
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

type OrderRepository struct{ mock.Mock }

var _ ordershared.Repository = (*OrderRepository)(nil)

// Create accepts a fixed order or a func(ctx, order) Order as return value.
func (m *OrderRepository) Create(ctx context.Context, order orderdomain.Order) (orderdomain.Order, error) {
	args := m.Called(ctx, order)
	if fn, ok := args.Get(0).(func(context.Context, orderdomain.Order) orderdomain.Order); ok {
		return fn(ctx, order), args.Error(1)
	}
	created, _ := args.Get(0).(orderdomain.Order)
	return created, args.Error(1)
}

func (m *OrderRepository) HasOrdersWithStatus(ctx context.Context, productID int, statuses []string) (bool, error) {
	args := m.Called(ctx, productID, statuses)
	return args.Bool(0), args.Error(1)
}

// UnitOfWork runs fn with the given mocks and records whether it would have
// committed: a real unit of work commits when fn returns nil.
type UnitOfWork struct {
	Carts     *Repository
	Products  *productmocks.Repository
	Orders    *OrderRepository
	Committed bool
}

func NewUnitOfWork() *UnitOfWork {
	return &UnitOfWork{
		Carts:    new(Repository),
		Products: new(productmocks.Repository),
		Orders:   new(OrderRepository),
	}
}

func (u *UnitOfWork) Do(_ context.Context, fn func(repos shared.TxRepositories) error) error {
	err := fn(shared.TxRepositories{Carts: u.Carts, Products: u.Products, Orders: u.Orders})
	u.Committed = err == nil
	return err
}
