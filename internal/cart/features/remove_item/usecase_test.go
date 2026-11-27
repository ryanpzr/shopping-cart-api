package removeitem_test

import (
	"context"
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

func TestRemoveItem(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByBuyer", mock.Anything, 7, domain.StatusActive).Return(domain.Cart{ID: 1, Items: []domain.Item{}}, nil)
	repo.On("RemoveItem", mock.Anything, 1, 10).Return(nil)

	resp, err := removeitem.NewUsecase(repo).RemoveItem(context.Background(), 7, 10)

	require.NoError(t, err)
	assert.Empty(t, resp.Items)
	repo.AssertExpectations(t)
}

func TestRemoveItem_NotInCart(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByBuyer", mock.Anything, 7, domain.StatusActive).Return(domain.Cart{ID: 1}, nil)
	repo.On("RemoveItem", mock.Anything, 1, 99).Return(domain.ErrItemNotFound)

	_, err := removeitem.NewUsecase(repo).RemoveItem(context.Background(), 7, 99)

	assertAppError(t, err, http.StatusNotFound)
}

func TestRemoveItem_NoActiveCart(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByBuyer", mock.Anything, 7, domain.StatusActive).Return(nil, domain.ErrCartNotFound)

	_, err := removeitem.NewUsecase(repo).RemoveItem(context.Background(), 7, 10)

	assertAppError(t, err, http.StatusNotFound)
	repo.AssertNotCalled(t, "RemoveItem", mock.Anything, mock.Anything, mock.Anything)
}
