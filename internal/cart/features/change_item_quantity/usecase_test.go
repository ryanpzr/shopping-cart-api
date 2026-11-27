package changeitemquantity_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/domain"
	changeitemquantity "github.com/ryanpzr/shopping-cart-api/internal/cart/features/change_item_quantity"
	"github.com/ryanpzr/shopping-cart-api/internal/cart/shared/mocks"
	productdomain "github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
	"github.com/ryanpzr/shopping-cart-api/pkg/money"
)

const buyer = 7

func ptr[T any](v T) *T { return &v }

func assertAppError(t *testing.T, err error, code int) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, code, appErr.Code)
}

var item = domain.Item{ID: 10, CartID: 1, ProductID: 5, Title: "Lamp", Quantity: 1, PriceSnapshot: money.FromCents(4500)}

func setup() *mocks.UnitOfWork {
	uow := mocks.NewUnitOfWork()
	uow.Carts.On("FindByBuyer", mock.Anything, buyer, domain.StatusActive).Return(domain.Cart{ID: 1, Items: []domain.Item{item}}, nil).Maybe()
	uow.Carts.On("LockByBuyer", mock.Anything, buyer, domain.StatusActive).Return(domain.Cart{ID: 1, Items: []domain.Item{item}}, nil)
	uow.Products.On("FindByID", mock.Anything, 5).Return(&productdomain.Product{ID: 5, Title: "Lamp", Quantity: 3}, nil).Maybe()
	return uow
}

func TestChangeQuantity_KeepsPriceSnapshot(t *testing.T) {
	uow := setup()
	want := item
	want.Quantity = 3
	uow.Carts.On("UpdateItem", mock.Anything, want).Return(nil)

	_, err := changeitemquantity.NewUsecase(uow).ChangeQuantity(context.Background(), buyer, 10,
		changeitemquantity.ChangeQuantityRequest{Quantity: ptr(3)})

	require.NoError(t, err)
	assert.True(t, uow.Committed)
	uow.Carts.AssertCalled(t, "FindByBuyer", mock.Anything, buyer, domain.StatusActive) // result read inside the transaction
	uow.Carts.AssertExpectations(t)
}

func TestChangeQuantity_AboveStock(t *testing.T) {
	uow := setup()

	_, err := changeitemquantity.NewUsecase(uow).ChangeQuantity(context.Background(), buyer, 10,
		changeitemquantity.ChangeQuantityRequest{Quantity: ptr(4)})

	assertAppError(t, err, http.StatusConflict)
	uow.Carts.AssertNotCalled(t, "UpdateItem", mock.Anything, mock.Anything)
}

func TestChangeQuantity_ItemNotInCart(t *testing.T) {
	uow := setup()

	_, err := changeitemquantity.NewUsecase(uow).ChangeQuantity(context.Background(), buyer, 99,
		changeitemquantity.ChangeQuantityRequest{Quantity: ptr(1)})

	assertAppError(t, err, http.StatusNotFound)
}

func TestChangeQuantity_NoActiveCart(t *testing.T) {
	uow := mocks.NewUnitOfWork()
	uow.Carts.On("LockByBuyer", mock.Anything, buyer, domain.StatusActive).Return(nil, domain.ErrCartNotFound)

	_, err := changeitemquantity.NewUsecase(uow).ChangeQuantity(context.Background(), buyer, 10,
		changeitemquantity.ChangeQuantityRequest{Quantity: ptr(1)})

	assertAppError(t, err, http.StatusNotFound)
}

func TestChangeQuantity_BadRequests(t *testing.T) {
	for name, q := range map[string]*int{"missing": nil, "zero": ptr(0), "negative": ptr(-1)} {
		t.Run(name, func(t *testing.T) {
			uow := mocks.NewUnitOfWork()
			_, err := changeitemquantity.NewUsecase(uow).ChangeQuantity(context.Background(), buyer, 10,
				changeitemquantity.ChangeQuantityRequest{Quantity: q})
			assertAppError(t, err, http.StatusBadRequest)
			uow.Carts.AssertNotCalled(t, "LockByBuyer", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}
