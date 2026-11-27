package removeitem_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/domain"
	removeitem "github.com/ryanpzr/shopping-cart-api/internal/cart/features/remove_item"
	"github.com/ryanpzr/shopping-cart-api/internal/cart/shared/mocks"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

func assertAppError(t *testing.T, err error, code int) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, code, appErr.Code)
}

func TestRemoveItem_LocksCartAndReadsResultInTheTransaction(t *testing.T) {
	uow := mocks.NewUnitOfWork()
	uow.Carts.On("LockByBuyer", mock.Anything, 7, domain.StatusActive).Return(domain.Cart{ID: 1}, nil)
	uow.Carts.On("RemoveItem", mock.Anything, 1, 10).Return(nil)
	uow.Carts.On("FindByBuyer", mock.Anything, 7, domain.StatusActive).Return(domain.Cart{ID: 1, Items: []domain.Item{}}, nil)

	resp, err := removeitem.NewUsecase(uow).RemoveItem(context.Background(), 7, 10)

	require.NoError(t, err)
	assert.Empty(t, resp.Items)
	assert.True(t, uow.Committed)
	uow.Carts.AssertExpectations(t)
}

func TestRemoveItem_NotInCart(t *testing.T) {
	uow := mocks.NewUnitOfWork()
	uow.Carts.On("LockByBuyer", mock.Anything, 7, domain.StatusActive).Return(domain.Cart{ID: 1}, nil)
	uow.Carts.On("RemoveItem", mock.Anything, 1, 99).Return(domain.ErrItemNotFound)

	_, err := removeitem.NewUsecase(uow).RemoveItem(context.Background(), 7, 99)

	assertAppError(t, err, http.StatusNotFound)
	assert.False(t, uow.Committed)
}

func TestRemoveItem_NoActiveCart(t *testing.T) {
	// Covers a cart closed by a concurrent checkout: nothing is deleted.
	uow := mocks.NewUnitOfWork()
	uow.Carts.On("LockByBuyer", mock.Anything, 7, domain.StatusActive).Return(nil, domain.ErrCartNotFound)

	_, err := removeitem.NewUsecase(uow).RemoveItem(context.Background(), 7, 10)

	assertAppError(t, err, http.StatusNotFound)
	uow.Carts.AssertNotCalled(t, "RemoveItem", mock.Anything, mock.Anything, mock.Anything)
}

func TestRemoveItem_RepositoryFailureRollsBack(t *testing.T) {
	uow := mocks.NewUnitOfWork()
	uow.Carts.On("LockByBuyer", mock.Anything, 7, domain.StatusActive).Return(domain.Cart{ID: 1}, nil)
	uow.Carts.On("RemoveItem", mock.Anything, 1, 10).Return(errors.New("connection reset"))

	_, err := removeitem.NewUsecase(uow).RemoveItem(context.Background(), 7, 10)

	assertAppError(t, err, http.StatusInternalServerError)
	assert.False(t, uow.Committed)
}
