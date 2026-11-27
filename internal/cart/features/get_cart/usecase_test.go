package getcart_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/domain"
	getcart "github.com/ryanpzr/shopping-cart-api/internal/cart/features/get_cart"
	"github.com/ryanpzr/shopping-cart-api/internal/cart/shared/mocks"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
	"github.com/ryanpzr/shopping-cart-api/pkg/money"
)

func TestGet_ExistingCartWithTotals(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByBuyer", mock.Anything, 7, domain.StatusActive).Return(domain.Cart{ID: 1, Status: domain.StatusActive, Items: []domain.Item{
		{ID: 10, ProductID: 5, Title: "Lamp", Quantity: 2, PriceSnapshot: money.FromCents(8990)},
	}}, nil)

	resp, err := getcart.NewUsecase(repo).Get(context.Background(), 7)

	require.NoError(t, err)
	assert.Equal(t, "179.80", resp.Subtotal.String())
	assert.Equal(t, "0.00", resp.DiscountTotal.String())
	assert.Equal(t, "179.80", resp.Total.String())
	assert.Equal(t, "Lamp", resp.Items[0].Title)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
}

func TestGet_CreatesCartOnFirstUse(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByBuyer", mock.Anything, 7, domain.StatusActive).Return(nil, domain.ErrCartNotFound)
	repo.On("Create", mock.Anything, 7, domain.StatusActive).Return(domain.Cart{ID: 2, Status: domain.StatusActive, Items: []domain.Item{}}, nil)

	resp, err := getcart.NewUsecase(repo).Get(context.Background(), 7)

	require.NoError(t, err)
	assert.Equal(t, 2, resp.ID)
	assert.NotNil(t, resp.Items)
	assert.Empty(t, resp.Items)
}

func TestGet_LostCreationRaceReadsTheOtherCart(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByBuyer", mock.Anything, 7, domain.StatusActive).Return(nil, domain.ErrCartNotFound).Once()
	repo.On("Create", mock.Anything, 7, domain.StatusActive).Return(nil, domain.ErrActiveCartExists)
	repo.On("FindByBuyer", mock.Anything, 7, domain.StatusActive).Return(domain.Cart{ID: 3, Items: []domain.Item{}}, nil).Once()

	resp, err := getcart.NewUsecase(repo).Get(context.Background(), 7)

	require.NoError(t, err)
	assert.Equal(t, 3, resp.ID)
}

func TestGet_RepositoryFailure(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByBuyer", mock.Anything, 7, domain.StatusActive).Return(nil, errors.New("connection refused"))

	_, err := getcart.NewUsecase(repo).Get(context.Background(), 7)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusInternalServerError, appErr.Code)
}
