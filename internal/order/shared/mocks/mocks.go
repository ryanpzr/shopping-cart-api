// Package mocks provides test doubles for the order repository and the order
// unit of work, shared by the order, cart and product feature tests.
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/ryanpzr/shopping-cart-api/internal/order/domain"
	"github.com/ryanpzr/shopping-cart-api/internal/order/shared"
	productmocks "github.com/ryanpzr/shopping-cart-api/internal/product/shared/mocks"
	outboxmocks "github.com/ryanpzr/shopping-cart-api/pkg/outbox/mocks"
)

type Repository struct{ mock.Mock }

var _ shared.Repository = (*Repository)(nil)

// Create accepts a fixed order or a func(ctx, order) Order as return value.
func (m *Repository) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	args := m.Called(ctx, order)
	if fn, ok := args.Get(0).(func(context.Context, domain.Order) domain.Order); ok {
		return fn(ctx, order), args.Error(1)
	}
	created, _ := args.Get(0).(domain.Order)
	return created, args.Error(1)
}

func (m *Repository) FindByBuyer(ctx context.Context, buyerID, limit, offset int) ([]domain.Order, int, error) {
	args := m.Called(ctx, buyerID, limit, offset)
	orders, _ := args.Get(0).([]domain.Order)
	return orders, args.Int(1), args.Error(2)
}

func (m *Repository) FindByID(ctx context.Context, id int) (domain.Order, error) {
	args := m.Called(ctx, id)
	order, _ := args.Get(0).(domain.Order)
	return order, args.Error(1)
}

func (m *Repository) LockByID(ctx context.Context, id int) (domain.Order, error) {
	args := m.Called(ctx, id)
	order, _ := args.Get(0).(domain.Order)
	return order, args.Error(1)
}

func (m *Repository) UpdateStatus(ctx context.Context, id int, status string) error {
	return m.Called(ctx, id, status).Error(0)
}

func (m *Repository) HasOrdersWithStatus(ctx context.Context, productID int, statuses []string) (bool, error) {
	args := m.Called(ctx, productID, statuses)
	return args.Bool(0), args.Error(1)
}

// UnitOfWork runs fn with the given mocks and records whether it would have
// committed: a real unit of work commits when fn returns nil.
type UnitOfWork struct {
	Orders    *Repository
	Products  *productmocks.Repository
	Outbox    *outboxmocks.Writer
	Committed bool
}

var _ shared.UnitOfWork = (*UnitOfWork)(nil)

func NewUnitOfWork() *UnitOfWork {
	return &UnitOfWork{Orders: new(Repository), Products: new(productmocks.Repository), Outbox: new(outboxmocks.Writer)}
}

func (u *UnitOfWork) Do(_ context.Context, fn func(repos shared.TxRepositories) error) error {
	err := fn(shared.TxRepositories{Orders: u.Orders, Products: u.Products, Outbox: u.Outbox})
	u.Committed = err == nil
	return err
}
