package listorders_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/order/domain"
	listorders "github.com/ryanpzr/shopping-cart-api/internal/order/features/list_orders"
	ordermocks "github.com/ryanpzr/shopping-cart-api/internal/order/shared/mocks"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
	"github.com/ryanpzr/shopping-cart-api/pkg/money"
)

func TestList_MapsOrdersItemsAndTotals(t *testing.T) {
	repo := new(ordermocks.Repository)
	repo.On("FindByBuyer", mock.Anything, 7, 20, 0).Return([]domain.Order{{
		ID: 55, CartID: 3, BuyerID: 7, Status: domain.StatusPending, TotalPrice: money.FromCents(38484),
		Items: []domain.Item{
			{ProductID: 1, Title: "Keyboard", Quantity: 2, UnitPrice: money.FromCents(19990), DiscountPercentage: 15},
			{ProductID: 2, Title: "Lamp", Quantity: 1, UnitPrice: money.FromCents(4500)},
		},
	}}, 1, nil)

	resp, err := listorders.NewUsecase(repo).List(context.Background(), 7, listorders.ListRequest{})

	require.NoError(t, err)
	require.Len(t, resp.Data, 1)
	order := resp.Data[0]
	assert.Equal(t, "384.84", order.TotalPrice.String())
	assert.Equal(t, "Keyboard", order.Items[0].Title)
	assert.Equal(t, "169.92", order.Items[0].SalePrice.String())
	assert.Equal(t, "339.84", order.Items[0].Subtotal.String())
	assert.Equal(t, listorders.MetaResponse{Total: 1, Limit: 20, Offset: 0}, resp.Meta)
}

func TestList_EmptyIsAnEmptyArray(t *testing.T) {
	repo := new(ordermocks.Repository)
	repo.On("FindByBuyer", mock.Anything, 7, mock.Anything, mock.Anything).Return(nil, 0, nil)

	resp, err := listorders.NewUsecase(repo).List(context.Background(), 7, listorders.ListRequest{})

	require.NoError(t, err)
	assert.NotNil(t, resp.Data)
	assert.Empty(t, resp.Data)
}

func TestList_Pagination(t *testing.T) {
	tests := []struct {
		name                  string
		limit, offset         int
		wantLimit, wantOffset int
	}{
		{"default limit", 0, 0, 20, 0},
		{"negative limit", -5, 0, 20, 0},
		{"custom page", 10, 30, 10, 30},
		{"limit capped", 500, 0, 100, 0},
		{"negative offset", 10, -1, 10, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(ordermocks.Repository)
			repo.On("FindByBuyer", mock.Anything, 7, tt.wantLimit, tt.wantOffset).Return([]domain.Order{}, 0, nil)

			resp, err := listorders.NewUsecase(repo).List(context.Background(), 7,
				listorders.ListRequest{Limit: tt.limit, Offset: tt.offset})

			require.NoError(t, err)
			assert.Equal(t, tt.wantLimit, resp.Meta.Limit)
			assert.Equal(t, tt.wantOffset, resp.Meta.Offset)
			repo.AssertExpectations(t)
		})
	}
}

func TestList_RepositoryFailure(t *testing.T) {
	repo := new(ordermocks.Repository)
	repo.On("FindByBuyer", mock.Anything, 7, mock.Anything, mock.Anything).Return(nil, 0, errors.New("connection refused"))

	_, err := listorders.NewUsecase(repo).List(context.Background(), 7, listorders.ListRequest{})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusInternalServerError, appErr.Code)
}
