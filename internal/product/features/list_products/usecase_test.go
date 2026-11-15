package listproducts_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	listproducts "github.com/ryanpzr/shopping-cart-api/internal/product/features/list_products"
	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
	"github.com/ryanpzr/shopping-cart-api/internal/product/shared/mocks"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

func ptr[T any](v T) *T { return &v }

// The public listing always asks the repository for active products only.
var activeOnly = shared.ProductFilters{Status: ptr(domain.StatusActive)}

func TestList_MapsProductsAndMeta(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindAll", mock.Anything, activeOnly, 20, 0).Return([]domain.Product{
		{ID: 1, Title: "A", Price: 100, DiscountPercentage: 20, Status: domain.StatusActive},
		{ID: 2, Title: "B", Price: 10, Status: domain.StatusActive},
	}, 2, nil)

	resp, err := listproducts.NewUsecase(repo).List(context.Background(), listproducts.ListRequest{Limit: 20})

	require.NoError(t, err)
	require.Len(t, resp.Data, 2)
	assert.Equal(t, 80.0, resp.Data[0].FinalPrice)
	assert.Equal(t, 10.0, resp.Data[1].FinalPrice)
	assert.Equal(t, listproducts.MetaResponse{Total: 2, Limit: 20, Offset: 0}, resp.Meta)
}

func TestList_EmptyResultIsEmptyArray(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, 0, nil)

	resp, err := listproducts.NewUsecase(repo).List(context.Background(), listproducts.ListRequest{})

	require.NoError(t, err)
	assert.NotNil(t, resp.Data)
	assert.Empty(t, resp.Data)
}

func TestList_PassesTrimmedFilters(t *testing.T) {
	repo := new(mocks.Repository)
	want := shared.ProductFilters{
		Status:   ptr(domain.StatusActive),
		Category: ptr("Books"),
		MinPrice: ptr(10.0),
		MaxPrice: ptr(50.0),
		Search:   ptr("go"),
	}
	repo.On("FindAll", mock.Anything, want, 20, 0).Return([]domain.Product{}, 0, nil)

	_, err := listproducts.NewUsecase(repo).List(context.Background(), listproducts.ListRequest{
		Category: ptr(" Books "), MinPrice: ptr(10.0), MaxPrice: ptr(50.0), Search: ptr("  go "),
	})

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestList_BlankFiltersAreIgnored(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindAll", mock.Anything, activeOnly, 20, 0).Return([]domain.Product{}, 0, nil)

	_, err := listproducts.NewUsecase(repo).List(context.Background(), listproducts.ListRequest{
		Category: ptr(""), Search: ptr("   "),
	})

	require.NoError(t, err)
	repo.AssertExpectations(t)
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
			repo := new(mocks.Repository)
			repo.On("FindAll", mock.Anything, mock.Anything, tt.wantLimit, tt.wantOffset).Return([]domain.Product{}, 0, nil)

			resp, err := listproducts.NewUsecase(repo).List(context.Background(),
				listproducts.ListRequest{Limit: tt.limit, Offset: tt.offset})

			require.NoError(t, err)
			assert.Equal(t, tt.wantLimit, resp.Meta.Limit)
			assert.Equal(t, tt.wantOffset, resp.Meta.Offset)
			repo.AssertExpectations(t)
		})
	}
}

func TestList_InvalidFilters(t *testing.T) {
	tests := []struct {
		name string
		req  listproducts.ListRequest
	}{
		{"negative min", listproducts.ListRequest{MinPrice: ptr(-1.0)}},
		{"negative max", listproducts.ListRequest{MaxPrice: ptr(-1.0)}},
		{"min above max", listproducts.ListRequest{MinPrice: ptr(50.0), MaxPrice: ptr(10.0)}},
		{"unknown category", listproducts.ListRequest{Category: ptr("books")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.Repository)

			_, err := listproducts.NewUsecase(repo).List(context.Background(), tt.req)

			var appErr *apperrors.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, http.StatusBadRequest, appErr.Code)
			repo.AssertNotCalled(t, "FindAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}
