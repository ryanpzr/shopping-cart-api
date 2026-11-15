package createproduct_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	createproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/create_product"
	"github.com/ryanpzr/shopping-cart-api/internal/product/shared/mocks"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

func ptr[T any](v T) *T { return &v }

func validRequest() createproduct.CreateRequest {
	return createproduct.CreateRequest{
		Title:    "Mechanical keyboard",
		Price:    ptr(200.0),
		Quantity: ptr(5),
	}
}

func assertAppError(t *testing.T, err error, code int) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, code, appErr.Code)
}

func TestCreate_Success(t *testing.T) {
	repo := new(mocks.Repository)
	req := validRequest()
	req.DiscountPercentage = ptr(25)
	req.Category = ptr("Electronics")

	repo.On("Create", mock.Anything, mock.MatchedBy(func(p domain.Product) bool {
		return p.SellerID == 7 && p.Title == "Mechanical keyboard" && p.Price == 200 &&
			p.DiscountPercentage == 25 && p.Quantity == 5 && p.Status == domain.StatusActive &&
			*p.Category == "Electronics"
	})).Return(domain.Product{
		ID: 1, SellerID: 7, Title: "Mechanical keyboard", Price: 200,
		DiscountPercentage: 25, Quantity: 5, Status: domain.StatusActive, Category: ptr("Electronics"),
	}, nil)

	resp, err := createproduct.NewUsecase(repo).Create(context.Background(), 7, req)

	require.NoError(t, err)
	assert.Equal(t, 1, resp.ID)
	assert.Equal(t, 7, resp.SellerID)
	assert.Equal(t, 150.0, resp.FinalPrice)
	assert.Equal(t, domain.StatusActive, resp.Status)
	repo.AssertExpectations(t)
}

func TestCreate_DefaultsDiscountToZeroAndAcceptsZeroQuantity(t *testing.T) {
	repo := new(mocks.Repository)
	req := validRequest()
	req.Quantity = ptr(0)

	repo.On("Create", mock.Anything, mock.MatchedBy(func(p domain.Product) bool {
		return p.DiscountPercentage == 0 && p.Quantity == 0
	})).Return(domain.Product{ID: 1, Price: 200, Status: domain.StatusActive}, nil)

	resp, err := createproduct.NewUsecase(repo).Create(context.Background(), 7, req)

	require.NoError(t, err)
	assert.Equal(t, 200.0, resp.FinalPrice)
	repo.AssertExpectations(t)
}

func TestCreate_ValidationErrors(t *testing.T) {
	tests := []struct {
		name   string
		modify func(r *createproduct.CreateRequest)
	}{
		{"missing title", func(r *createproduct.CreateRequest) { r.Title = "" }},
		{"blank title", func(r *createproduct.CreateRequest) { r.Title = "   " }},
		{"title too long", func(r *createproduct.CreateRequest) { r.Title = strings.Repeat("a", 256) }},
		{"missing price", func(r *createproduct.CreateRequest) { r.Price = nil }},
		{"zero price", func(r *createproduct.CreateRequest) { r.Price = ptr(0.0) }},
		{"negative price", func(r *createproduct.CreateRequest) { r.Price = ptr(-10.0) }},
		{"price above limit", func(r *createproduct.CreateRequest) { r.Price = ptr(1_000_000.01) }},
		{"missing quantity", func(r *createproduct.CreateRequest) { r.Quantity = nil }},
		{"negative quantity", func(r *createproduct.CreateRequest) { r.Quantity = ptr(-1) }},
		{"negative discount", func(r *createproduct.CreateRequest) { r.DiscountPercentage = ptr(-1) }},
		{"discount above 80", func(r *createproduct.CreateRequest) { r.DiscountPercentage = ptr(81) }},
		{"unknown category", func(r *createproduct.CreateRequest) { r.Category = ptr("Weapons") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.Repository)
			req := validRequest()
			tt.modify(&req)

			_, err := createproduct.NewUsecase(repo).Create(context.Background(), 7, req)

			assertAppError(t, err, http.StatusBadRequest)
			repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		})
	}
}

func TestCreate_RepositoryError(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("Create", mock.Anything, mock.Anything).
		Return(domain.Product{}, errors.New("connection refused"))

	_, err := createproduct.NewUsecase(repo).Create(context.Background(), 7, validRequest())

	assertAppError(t, err, http.StatusInternalServerError)
}
