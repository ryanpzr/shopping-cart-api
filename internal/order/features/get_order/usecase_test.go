package getorder_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/order/domain"
	getorder "github.com/ryanpzr/shopping-cart-api/internal/order/features/get_order"
	ordermocks "github.com/ryanpzr/shopping-cart-api/internal/order/shared/mocks"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
	"github.com/ryanpzr/shopping-cart-api/pkg/money"
)

func assertAppError(t *testing.T, err error, code int) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, code, appErr.Code)
}

func order() domain.Order {
	return domain.Order{ID: 55, BuyerID: 7, Status: domain.StatusPaid, TotalPrice: money.FromCents(4500), Items: []domain.Item{
		{ProductID: 2, Title: "Lamp", Quantity: 1, UnitPrice: money.FromCents(4500)},
	}}
}

func TestGet_OwnerAndAdminCanRead(t *testing.T) {
	tests := []struct {
		name   string
		userID int
		role   string
	}{
		{"owner", 7, "client"},
		{"admin reads another user's order", 1, "admin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(ordermocks.Repository)
			repo.On("FindByID", mock.Anything, 55).Return(order(), nil)

			resp, err := getorder.NewUsecase(repo).Get(context.Background(), 55, tt.userID, tt.role)

			require.NoError(t, err)
			assert.Equal(t, 55, resp.ID)
			assert.Equal(t, "Lamp", resp.Items[0].Title)
			assert.Equal(t, "45.00", resp.TotalPrice.String())
		})
	}
}

func TestGet_AnotherClientIsForbidden(t *testing.T) {
	repo := new(ordermocks.Repository)
	repo.On("FindByID", mock.Anything, 55).Return(order(), nil)

	_, err := getorder.NewUsecase(repo).Get(context.Background(), 55, 8, "client")

	assertAppError(t, err, http.StatusForbidden)
}

func TestGet_NotFound(t *testing.T) {
	repo := new(ordermocks.Repository)
	repo.On("FindByID", mock.Anything, 55).Return(nil, domain.ErrOrderNotFound)

	_, err := getorder.NewUsecase(repo).Get(context.Background(), 55, 7, "client")

	assertAppError(t, err, http.StatusNotFound)
}

func TestGet_NotFoundIsReportedBeforeOwnership(t *testing.T) {
	// A missing order is a 404 for everyone, including a client who would not own it.
	repo := new(ordermocks.Repository)
	repo.On("FindByID", mock.Anything, 99).Return(nil, domain.ErrOrderNotFound)

	_, err := getorder.NewUsecase(repo).Get(context.Background(), 99, 8, "client")

	assertAppError(t, err, http.StatusNotFound)
}

func TestGet_RepositoryFailure(t *testing.T) {
	repo := new(ordermocks.Repository)
	repo.On("FindByID", mock.Anything, 55).Return(nil, errors.New("connection refused"))

	_, err := getorder.NewUsecase(repo).Get(context.Background(), 55, 7, "client")

	assertAppError(t, err, http.StatusInternalServerError)
}
