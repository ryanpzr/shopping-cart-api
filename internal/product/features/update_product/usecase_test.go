package updateproduct_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	updateproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/update_product"
	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
	"github.com/ryanpzr/shopping-cart-api/internal/product/shared/mocks"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

func ptr[T any](v T) *T { return &v }

func assertAppError(t *testing.T, err error, code int) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, code, appErr.Code)
}

var owned = &domain.Product{ID: 3, SellerID: 7, Title: "Old", Price: 100, Quantity: 1, Status: domain.StatusActive}

func TestUpdate_PartialUpdateByOwner(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(owned, nil)
	repo.On("Update", mock.Anything, 3, shared.ProductUpdate{
		Price: ptr(80.0), DiscountPercentage: ptr(50),
	}).Return(domain.Product{ID: 3, SellerID: 7, Title: "Old", Price: 80, DiscountPercentage: 50, Status: domain.StatusActive}, nil)

	resp, err := updateproduct.NewUsecase(repo).Update(context.Background(), 3, 7,
		updateproduct.UpdateRequest{Price: ptr(80.0), DiscountPercentage: ptr(50)})

	require.NoError(t, err)
	assert.Equal(t, "Old", resp.Title)
	assert.Equal(t, 40.0, resp.FinalPrice)
	repo.AssertExpectations(t)
}

func TestUpdate_NotOwner(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(owned, nil)

	_, err := updateproduct.NewUsecase(repo).Update(context.Background(), 3, 8,
		updateproduct.UpdateRequest{Title: ptr("New")})

	assertAppError(t, err, http.StatusForbidden)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}

func TestUpdate_NotFound(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(nil, apperrors.NewNotFound("product not found"))

	_, err := updateproduct.NewUsecase(repo).Update(context.Background(), 3, 7,
		updateproduct.UpdateRequest{Title: ptr("New")})

	assertAppError(t, err, http.StatusNotFound)
}

func TestUpdate_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		req  updateproduct.UpdateRequest
	}{
		{"empty body", updateproduct.UpdateRequest{}},
		{"blank title", updateproduct.UpdateRequest{Title: ptr(" ")}},
		{"zero price", updateproduct.UpdateRequest{Price: ptr(0.0)}},
		{"price above limit", updateproduct.UpdateRequest{Price: ptr(2_000_000.0)}},
		{"discount above 80", updateproduct.UpdateRequest{DiscountPercentage: ptr(81)}},
		{"unknown category", updateproduct.UpdateRequest{Category: ptr("books")}},
		{"negative quantity", updateproduct.UpdateRequest{Quantity: ptr(-3)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.Repository)

			_, err := updateproduct.NewUsecase(repo).Update(context.Background(), 3, 7, tt.req)

			assertAppError(t, err, http.StatusBadRequest)
			repo.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
		})
	}
}

func TestUpdate_AcceptsZeroQuantity(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(owned, nil)
	repo.On("Update", mock.Anything, 3, shared.ProductUpdate{Quantity: ptr(0)}).
		Return(domain.Product{ID: 3, SellerID: 7, Price: 100, Quantity: 0}, nil)

	resp, err := updateproduct.NewUsecase(repo).Update(context.Background(), 3, 7,
		updateproduct.UpdateRequest{Quantity: ptr(0)})

	require.NoError(t, err)
	assert.Equal(t, 0, resp.Quantity)
}
