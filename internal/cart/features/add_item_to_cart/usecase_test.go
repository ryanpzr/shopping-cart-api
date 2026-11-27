package additemtocart_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/domain"
	additemtocart "github.com/ryanpzr/shopping-cart-api/internal/cart/features/add_item_to_cart"
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

// Lamp: 50.00 with 10% off = 45.00, 4 units in stock, sold by user 2.
func lamp() *productdomain.Product {
	return &productdomain.Product{ID: 5, SellerID: 2, Title: "Lamp", Price: 50, DiscountPercentage: 10, Quantity: 4, Status: productdomain.StatusActive}
}

type fixture struct {
	repo *mocks.Repository
	uow  *mocks.UnitOfWork
	uc   additemtocart.Usecase
}

func newFixture(c domain.Cart, product *productdomain.Product) fixture {
	repo := new(mocks.Repository)
	repo.On("FindByBuyer", mock.Anything, buyer, domain.StatusActive).Return(c, nil)
	uow := mocks.NewUnitOfWork()
	uow.Carts.On("LockByBuyer", mock.Anything, buyer, domain.StatusActive).Return(c, nil)
	uow.Products.On("FindByID", mock.Anything, 5).Return(product, nil)
	return fixture{repo: repo, uow: uow, uc: additemtocart.NewUsecase(repo, uow)}
}

func TestAddItem_NewItemUsesCurrentSalePrice(t *testing.T) {
	f := newFixture(domain.Cart{ID: 1, Items: []domain.Item{}}, lamp())
	f.uow.Carts.On("AddItem", mock.Anything, domain.Item{CartID: 1, ProductID: 5, Quantity: 3, PriceSnapshot: money.FromCents(4500)}).
		Return(domain.Item{ID: 10}, nil)

	_, err := f.uc.AddItem(context.Background(), buyer, additemtocart.AddItemRequest{ProductID: ptr(5), Quantity: ptr(3)})

	require.NoError(t, err)
	assert.True(t, f.uow.Committed)
	f.uow.Carts.AssertExpectations(t)
}

func TestAddItem_SameProductSumsQuantityAndRefreshesPrice(t *testing.T) {
	existing := domain.Item{ID: 10, CartID: 1, ProductID: 5, Quantity: 1, PriceSnapshot: money.FromCents(5000)}
	f := newFixture(domain.Cart{ID: 1, Items: []domain.Item{existing}}, lamp())
	f.uow.Carts.On("UpdateItem", mock.Anything, domain.Item{ID: 10, CartID: 1, ProductID: 5, Quantity: 4, PriceSnapshot: money.FromCents(4500)}).
		Return(nil)

	_, err := f.uc.AddItem(context.Background(), buyer, additemtocart.AddItemRequest{ProductID: ptr(5), Quantity: ptr(3)})

	require.NoError(t, err)
	f.uow.Carts.AssertExpectations(t)
	f.uow.Carts.AssertNotCalled(t, "AddItem", mock.Anything, mock.Anything)
}

func TestAddItem_CombinedQuantityAboveStock(t *testing.T) {
	existing := domain.Item{ID: 10, CartID: 1, ProductID: 5, Quantity: 3}
	f := newFixture(domain.Cart{ID: 1, Items: []domain.Item{existing}}, lamp())

	_, err := f.uc.AddItem(context.Background(), buyer, additemtocart.AddItemRequest{ProductID: ptr(5), Quantity: ptr(2)})

	assertAppError(t, err, http.StatusConflict)
	assert.False(t, f.uow.Committed)
}

func TestAddItem_ProductRules(t *testing.T) {
	inactive := lamp()
	inactive.Status = productdomain.StatusInactive
	own := lamp()
	own.SellerID = buyer

	for name, tt := range map[string]struct {
		product *productdomain.Product
		err     error
		code    int
	}{
		"missing product":  {nil, productdomain.ErrProductNotFound, http.StatusNotFound},
		"inactive product": {inactive, nil, http.StatusNotFound},
		"own product":      {own, nil, http.StatusForbidden},
	} {
		t.Run(name, func(t *testing.T) {
			repo := new(mocks.Repository)
			repo.On("FindByBuyer", mock.Anything, buyer, domain.StatusActive).Return(domain.Cart{ID: 1}, nil)
			uow := mocks.NewUnitOfWork()
			uow.Carts.On("LockByBuyer", mock.Anything, buyer, domain.StatusActive).Return(domain.Cart{ID: 1}, nil)
			uow.Products.On("FindByID", mock.Anything, 5).Return(tt.product, tt.err)

			_, err := additemtocart.NewUsecase(repo, uow).AddItem(context.Background(), buyer,
				additemtocart.AddItemRequest{ProductID: ptr(5), Quantity: ptr(1)})

			assertAppError(t, err, tt.code)
			assert.False(t, uow.Committed)
		})
	}
}

func TestAddItem_BadRequests(t *testing.T) {
	for name, req := range map[string]additemtocart.AddItemRequest{
		"missing product_id": {Quantity: ptr(1)},
		"missing quantity":   {ProductID: ptr(5)},
		"zero quantity":      {ProductID: ptr(5), Quantity: ptr(0)},
		"negative quantity":  {ProductID: ptr(5), Quantity: ptr(-2)},
	} {
		t.Run(name, func(t *testing.T) {
			repo := new(mocks.Repository)
			_, err := additemtocart.NewUsecase(repo, mocks.NewUnitOfWork()).AddItem(context.Background(), buyer, req)
			assertAppError(t, err, http.StatusBadRequest)
			repo.AssertNotCalled(t, "FindByBuyer", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestAddItem_CreatesCartOnFirstUse(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByBuyer", mock.Anything, buyer, domain.StatusActive).Return(nil, domain.ErrCartNotFound).Once()
	repo.On("Create", mock.Anything, buyer, domain.StatusActive).Return(domain.Cart{ID: 1}, nil)
	repo.On("FindByBuyer", mock.Anything, buyer, domain.StatusActive).Return(domain.Cart{ID: 1}, nil)
	uow := mocks.NewUnitOfWork()
	uow.Carts.On("LockByBuyer", mock.Anything, buyer, domain.StatusActive).Return(domain.Cart{ID: 1}, nil)
	uow.Products.On("FindByID", mock.Anything, 5).Return(lamp(), nil)
	uow.Carts.On("AddItem", mock.Anything, mock.Anything).Return(domain.Item{ID: 10}, nil)

	_, err := additemtocart.NewUsecase(repo, uow).AddItem(context.Background(), buyer,
		additemtocart.AddItemRequest{ProductID: ptr(5), Quantity: ptr(1)})

	require.NoError(t, err)
	repo.AssertExpectations(t)
}
