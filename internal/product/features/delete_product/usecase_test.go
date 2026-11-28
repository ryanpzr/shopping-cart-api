package deleteproduct_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	ordermocks "github.com/ryanpzr/shopping-cart-api/internal/order/shared/mocks"
	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	deleteproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/delete_product"
	"github.com/ryanpzr/shopping-cart-api/internal/product/shared/mocks"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

// fakeUnitOfWork runs fn with mocks and records whether it would have
// committed: a real unit of work commits when fn returns nil.
type fakeUnitOfWork struct {
	Products  *mocks.Repository
	Orders    *ordermocks.Repository
	Committed bool
}

func newUnitOfWork() *fakeUnitOfWork {
	return &fakeUnitOfWork{Products: new(mocks.Repository), Orders: new(ordermocks.Repository)}
}

func (u *fakeUnitOfWork) Do(_ context.Context, fn func(repos deleteproduct.TxRepositories) error) error {
	err := fn(deleteproduct.TxRepositories{Products: u.Products, Orders: u.Orders})
	u.Committed = err == nil
	return err
}

func assertAppError(t *testing.T, err error, code int) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, code, appErr.Code)
}

var product = &domain.Product{ID: 3, SellerID: 7, Title: "Desk", Status: domain.StatusActive}

func TestDelete_Allowed(t *testing.T) {
	tests := []struct {
		name   string
		userID int
		role   string
	}{
		{"owner", 7, "client"},
		{"admin deletes someone else's product", 1, "admin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uow := newUnitOfWork()
			uow.Products.On("LockByID", mock.Anything, 3).Return(product, nil)
			uow.Orders.On("HasOrdersWithStatus", mock.Anything, 3, domain.ActiveOrderStatuses).Return(false, nil)
			uow.Products.On("Delete", mock.Anything, 3).Return(nil)

			err := deleteproduct.NewUsecase(uow).Delete(context.Background(), 3, tt.userID, tt.role)

			require.NoError(t, err)
			assert.True(t, uow.Committed)
			uow.Products.AssertExpectations(t)
			uow.Orders.AssertExpectations(t)
		})
	}
}

func TestDelete_LocksTheProductBeforeLookingAtOrders(t *testing.T) {
	// The check and the delete only close the race with checkout if the
	// product row is locked first: assert the call order.
	var calls []string
	uow := newUnitOfWork()
	uow.Products.On("LockByID", mock.Anything, 3).Run(func(mock.Arguments) { calls = append(calls, "lock") }).Return(product, nil)
	uow.Orders.On("HasOrdersWithStatus", mock.Anything, 3, mock.Anything).Run(func(mock.Arguments) { calls = append(calls, "check") }).Return(false, nil)
	uow.Products.On("Delete", mock.Anything, 3).Run(func(mock.Arguments) { calls = append(calls, "delete") }).Return(nil)

	require.NoError(t, deleteproduct.NewUsecase(uow).Delete(context.Background(), 3, 7, "client"))

	assert.Equal(t, []string{"lock", "check", "delete"}, calls)
}

func TestDelete_NonOwnerClient(t *testing.T) {
	uow := newUnitOfWork()
	uow.Products.On("LockByID", mock.Anything, 3).Return(product, nil)

	err := deleteproduct.NewUsecase(uow).Delete(context.Background(), 3, 8, "client")

	assertAppError(t, err, http.StatusForbidden)
	assert.False(t, uow.Committed)
	uow.Products.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
	uow.Orders.AssertNotCalled(t, "HasOrdersWithStatus", mock.Anything, mock.Anything, mock.Anything)
}

func TestDelete_ActiveOrdersConflict(t *testing.T) {
	uow := newUnitOfWork()
	uow.Products.On("LockByID", mock.Anything, 3).Return(product, nil)
	uow.Orders.On("HasOrdersWithStatus", mock.Anything, 3, []string{"pending", "paid", "shipped"}).Return(true, nil)

	err := deleteproduct.NewUsecase(uow).Delete(context.Background(), 3, 7, "client")

	assertAppError(t, err, http.StatusConflict)
	assert.False(t, uow.Committed)
	uow.Products.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestDelete_NotFound(t *testing.T) {
	uow := newUnitOfWork()
	uow.Products.On("LockByID", mock.Anything, 3).Return(nil, domain.ErrProductNotFound)

	err := deleteproduct.NewUsecase(uow).Delete(context.Background(), 3, 7, "client")

	assertAppError(t, err, http.StatusNotFound)
}

func TestDelete_ProductStillReferencedByOrders(t *testing.T) {
	// No active order, but a delivered or cancelled one still references the
	// product: the foreign key refuses the delete and the answer is 409.
	uow := newUnitOfWork()
	uow.Products.On("LockByID", mock.Anything, 3).Return(product, nil)
	uow.Orders.On("HasOrdersWithStatus", mock.Anything, 3, mock.Anything).Return(false, nil)
	uow.Products.On("Delete", mock.Anything, 3).Return(domain.ErrProductInUse)

	err := deleteproduct.NewUsecase(uow).Delete(context.Background(), 3, 7, "client")

	assertAppError(t, err, http.StatusConflict)
	assert.False(t, uow.Committed)
}

func TestDelete_OrderCheckFails(t *testing.T) {
	uow := newUnitOfWork()
	uow.Products.On("LockByID", mock.Anything, 3).Return(product, nil)
	uow.Orders.On("HasOrdersWithStatus", mock.Anything, 3, mock.Anything).Return(false, errors.New("connection refused"))

	err := deleteproduct.NewUsecase(uow).Delete(context.Background(), 3, 7, "client")

	assertAppError(t, err, http.StatusInternalServerError)
	assert.False(t, uow.Committed)
	uow.Products.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}
