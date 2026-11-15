package deleteproduct_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	deleteproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/delete_product"
	"github.com/ryanpzr/shopping-cart-api/internal/product/shared/mocks"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

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
			repo := new(mocks.Repository)
			repo.On("FindByID", mock.Anything, 3).Return(product, nil)
			repo.On("HasActiveOrders", mock.Anything, 3).Return(false, nil)
			repo.On("Delete", mock.Anything, 3).Return(nil)

			err := deleteproduct.NewUsecase(repo).Delete(context.Background(), 3, tt.userID, tt.role)

			require.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestDelete_NonOwnerClient(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(product, nil)

	err := deleteproduct.NewUsecase(repo).Delete(context.Background(), 3, 8, "client")

	assertAppError(t, err, http.StatusForbidden)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestDelete_ActiveOrdersConflict(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(product, nil)
	repo.On("HasActiveOrders", mock.Anything, 3).Return(true, nil)

	err := deleteproduct.NewUsecase(repo).Delete(context.Background(), 3, 7, "client")

	assertAppError(t, err, http.StatusConflict)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestDelete_NotFound(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(nil, apperrors.NewNotFound("product not found"))

	err := deleteproduct.NewUsecase(repo).Delete(context.Background(), 3, 7, "client")

	assertAppError(t, err, http.StatusNotFound)
}

func TestDelete_OrderCheckFails(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(product, nil)
	repo.On("HasActiveOrders", mock.Anything, 3).Return(false, apperrors.NewInternalServer("db down"))

	err := deleteproduct.NewUsecase(repo).Delete(context.Background(), 3, 7, "client")

	assertAppError(t, err, http.StatusInternalServerError)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}
