// Package mocks provides a testify mock of the product repository, shared by
// the feature tests.
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
)

type Repository struct{ mock.Mock }

var _ shared.Repository = (*Repository)(nil)

func (m *Repository) FindAll(ctx context.Context, f shared.ProductFilters, limit, offset int) ([]domain.Product, int, error) {
	args := m.Called(ctx, f, limit, offset)
	products, _ := args.Get(0).([]domain.Product)
	return products, args.Int(1), args.Error(2)
}

func (m *Repository) FindByID(ctx context.Context, id int) (*domain.Product, error) {
	args := m.Called(ctx, id)
	p, _ := args.Get(0).(*domain.Product)
	return p, args.Error(1)
}

func (m *Repository) Create(ctx context.Context, p domain.Product) (domain.Product, error) {
	args := m.Called(ctx, p)
	created, _ := args.Get(0).(domain.Product)
	return created, args.Error(1)
}

func (m *Repository) Update(ctx context.Context, id int, u shared.ProductUpdate) (domain.Product, error) {
	args := m.Called(ctx, id, u)
	updated, _ := args.Get(0).(domain.Product)
	return updated, args.Error(1)
}

func (m *Repository) UpdateStatus(ctx context.Context, id int, status string) error {
	return m.Called(ctx, id, status).Error(0)
}

func (m *Repository) Delete(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}

func (m *Repository) HasActiveOrders(ctx context.Context, productID int) (bool, error) {
	args := m.Called(ctx, productID)
	return args.Bool(0), args.Error(1)
}
