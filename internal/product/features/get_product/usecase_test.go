package getproduct_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	getproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/get_product"
	"github.com/ryanpzr/shopping-cart-api/internal/product/shared/mocks"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

func assertAppError(t *testing.T, err error, code int) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, code, appErr.Code)
}

func TestGet_ActiveProduct(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(&domain.Product{
		ID: 3, SellerID: 1, Title: "Desk", Price: 500, DiscountPercentage: 10, Quantity: 4, Status: domain.StatusActive,
	}, nil)

	resp, err := getproduct.NewUsecase(repo).Get(context.Background(), 3)

	require.NoError(t, err)
	assert.Equal(t, 3, resp.ID)
	assert.Equal(t, 450.0, resp.FinalPrice)
}

func TestGet_InactiveProductIsNotFound(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(&domain.Product{ID: 3, Status: domain.StatusInactive}, nil)

	_, err := getproduct.NewUsecase(repo).Get(context.Background(), 3)

	assertAppError(t, err, http.StatusNotFound)
}

func TestGet_NotFound(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(nil, apperrors.NewNotFound("product not found"))

	_, err := getproduct.NewUsecase(repo).Get(context.Background(), 3)

	assertAppError(t, err, http.StatusNotFound)
}
