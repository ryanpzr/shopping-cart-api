package updateproduct_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	updateproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/update_product"
	"github.com/ryanpzr/shopping-cart-api/internal/product/shared/mocks"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

func ptr[T any](v T) *T { return &v }

func set[T any](v T) updateproduct.Optional[T] {
	return updateproduct.Optional[T]{Set: true, Value: &v}
}

func null[T any]() updateproduct.Optional[T] {
	return updateproduct.Optional[T]{Set: true}
}

func assertAppError(t *testing.T, err error, code int) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, code, appErr.Code)
}

func owned() *domain.Product {
	return &domain.Product{
		ID: 3, SellerID: 7, Title: "Old", Price: 100, Quantity: 1, Status: domain.StatusActive,
		Photo: ptr("https://cdn.example.com/a.jpg"), Description: ptr("desc"), Category: ptr("Books"),
	}
}

// echo makes the repository return the product it was asked to save.
func echo(repo *mocks.Repository) {
	repo.On("Update", mock.Anything, mock.Anything).Return(func(_ context.Context, p domain.Product) domain.Product {
		return p
	}, nil)
}

func TestUpdate_PartialUpdateKeepsOtherFields(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(owned(), nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(p domain.Product) bool {
		return p.ID == 3 && p.Price == 80 && p.DiscountPercentage == 50 &&
			p.Title == "Old" && p.Quantity == 1 && *p.Category == "Books" && *p.Description == "desc"
	})).Return(domain.Product{ID: 3, SellerID: 7, Title: "Old", Price: 80, DiscountPercentage: 50}, nil)

	resp, err := updateproduct.NewUsecase(repo).Update(context.Background(), 3, 7,
		updateproduct.UpdateRequest{Price: set(80.0), DiscountPercentage: set(50)})

	require.NoError(t, err)
	assert.Equal(t, 40.0, resp.FinalPrice)
	repo.AssertExpectations(t)
}

func TestUpdate_NullClearsOptionalFields(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(owned(), nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(p domain.Product) bool {
		return p.Photo == nil && p.Description == nil && p.Category == nil && p.Title == "Old"
	})).Return(domain.Product{ID: 3, Title: "Old"}, nil)

	resp, err := updateproduct.NewUsecase(repo).Update(context.Background(), 3, 7, updateproduct.UpdateRequest{
		Photo: null[string](), Description: null[string](), Category: null[string](),
	})

	require.NoError(t, err)
	assert.Nil(t, resp.Category)
	repo.AssertExpectations(t)
}

func TestUpdate_ValidatesTheResultingProduct(t *testing.T) {
	// The stored product already breaks no rule; the merged result must not either.
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(owned(), nil)

	_, err := updateproduct.NewUsecase(repo).Update(context.Background(), 3, 7,
		updateproduct.UpdateRequest{Category: set("books")})

	assertAppError(t, err, http.StatusBadRequest)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestUpdate_AcceptsZeroQuantity(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(owned(), nil)
	echo(repo)

	resp, err := updateproduct.NewUsecase(repo).Update(context.Background(), 3, 7,
		updateproduct.UpdateRequest{Quantity: set(0)})

	require.NoError(t, err)
	assert.Equal(t, 0, resp.Quantity)
	assert.Equal(t, "Old", resp.Title, "the saved product is echoed back, not a zero value")
	assert.Equal(t, 100.0, resp.Price)
}

func TestUpdate_NotOwner(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(owned(), nil)

	_, err := updateproduct.NewUsecase(repo).Update(context.Background(), 3, 8,
		updateproduct.UpdateRequest{Title: set("New")})

	assertAppError(t, err, http.StatusForbidden)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestUpdate_NotFound(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(nil, domain.ErrProductNotFound)

	_, err := updateproduct.NewUsecase(repo).Update(context.Background(), 3, 7,
		updateproduct.UpdateRequest{Title: set("New")})

	assertAppError(t, err, http.StatusNotFound)
}

func TestUpdate_RepositoryFailure(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(nil, errors.New("connection refused"))

	_, err := updateproduct.NewUsecase(repo).Update(context.Background(), 3, 7,
		updateproduct.UpdateRequest{Title: set("New")})

	assertAppError(t, err, http.StatusInternalServerError)
}

func TestUpdate_BadRequests(t *testing.T) {
	tests := []struct {
		name string
		req  updateproduct.UpdateRequest
	}{
		{"empty body", updateproduct.UpdateRequest{}},
		{"null title", updateproduct.UpdateRequest{Title: null[string]()}},
		{"null price", updateproduct.UpdateRequest{Price: null[float64]()}},
		{"null discount", updateproduct.UpdateRequest{DiscountPercentage: null[int]()}},
		{"null quantity", updateproduct.UpdateRequest{Quantity: null[int]()}},
		{"blank title", updateproduct.UpdateRequest{Title: set(" ")}},
		{"zero price", updateproduct.UpdateRequest{Price: set(0.0)}},
		{"price above limit", updateproduct.UpdateRequest{Price: set(2_000_000.0)}},
		{"discount above 80", updateproduct.UpdateRequest{DiscountPercentage: set(81)}},
		{"unknown category", updateproduct.UpdateRequest{Category: set("books")}},
		{"negative quantity", updateproduct.UpdateRequest{Quantity: set(-3)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.Repository)
			repo.On("FindByID", mock.Anything, 3).Return(owned(), nil).Maybe()

			_, err := updateproduct.NewUsecase(repo).Update(context.Background(), 3, 7, tt.req)

			assertAppError(t, err, http.StatusBadRequest)
			repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
		})
	}
}
