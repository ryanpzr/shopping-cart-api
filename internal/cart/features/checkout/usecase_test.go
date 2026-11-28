package checkout_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/domain"
	"github.com/ryanpzr/shopping-cart-api/internal/cart/features/checkout"
	"github.com/ryanpzr/shopping-cart-api/internal/cart/shared/mocks"
	orderdomain "github.com/ryanpzr/shopping-cart-api/internal/order/domain"
	productdomain "github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
	"github.com/ryanpzr/shopping-cart-api/pkg/money"
	"github.com/ryanpzr/shopping-cart-api/pkg/outbox"
	outboxmocks "github.com/ryanpzr/shopping-cart-api/pkg/outbox/mocks"
)

const buyer = 7

func assertAppError(t *testing.T, err error, code int) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, code, appErr.Code)
}

// Keyboard: 199.90 with 15% off = 169.92. Lamp: 45.00, no discount.
func products() []productdomain.Product {
	return []productdomain.Product{
		{ID: 1, SellerID: 2, Title: "Keyboard", Price: 199.90, DiscountPercentage: 15, Quantity: 5, Status: productdomain.StatusActive},
		{ID: 2, SellerID: 3, Title: "Lamp", Price: 45, Quantity: 1, Status: productdomain.StatusActive},
	}
}

func cart() domain.Cart {
	return domain.Cart{ID: 10, BuyerID: buyer, Status: domain.StatusActive, Items: []domain.Item{
		{ID: 100, CartID: 10, ProductID: 1, Title: "Keyboard", Quantity: 2, PriceSnapshot: money.FromCents(16992)},
		{ID: 101, CartID: 10, ProductID: 2, Title: "Lamp", Quantity: 1, PriceSnapshot: money.FromCents(4500)},
	}}
}

func setup(c domain.Cart, ps []productdomain.Product) *mocks.UnitOfWork {
	uow := mocks.NewUnitOfWork()
	uow.Carts.On("LockByBuyer", mock.Anything, buyer, domain.StatusActive).Return(c, nil)
	uow.Products.On("LockByIDs", mock.Anything, c.ProductIDs()).Return(ps, nil).Maybe()
	return uow
}

func TestCheckout_PlacesOrder(t *testing.T) {
	uow := setup(cart(), products())
	uow.Products.On("UpdateStock", mock.Anything, 1, 3).Return(nil)
	uow.Products.On("UpdateStock", mock.Anything, 2, 0).Return(nil)
	uow.Orders.On("Create", mock.Anything, mock.MatchedBy(func(o orderdomain.Order) bool {
		return o.CartID == 10 && o.BuyerID == buyer && o.Status == orderdomain.StatusPending &&
			o.TotalPrice == money.FromCents(16992*2+4500) && len(o.Items) == 2 &&
			o.Items[0].UnitPrice == money.FromCents(19990) && o.Items[0].DiscountPercentage == 15
	})).Return(func(_ context.Context, o orderdomain.Order) orderdomain.Order {
		o.ID = 55
		o.CreatedAt = time.Now()
		return o
	}, nil)
	uow.Carts.On("UpdateStatus", mock.Anything, 10, domain.StatusCheckedOut).Return(nil)
	uow.Outbox.On("Insert", mock.Anything, outboxmocks.Matches(outbox.EventCartCheckout, buyer,
		map[string]any{"order_id": 55, "total_price": money.FromCents(38484)})).Return(nil)
	uow.Outbox.On("Insert", mock.Anything, outboxmocks.Matches(outbox.EventOrderPlaced, buyer,
		map[string]any{"order_id": 55})).Return(nil)

	resp, err := checkout.NewUsecase(uow).Checkout(context.Background(), buyer)

	require.NoError(t, err)
	assert.True(t, uow.Committed)
	assert.Equal(t, 55, resp.ID)
	assert.Equal(t, "384.84", resp.TotalPrice.String())
	assert.Equal(t, "169.92", resp.Items[0].SalePrice.String())
	assert.Equal(t, "339.84", resp.Items[0].Subtotal.String())
	uow.Products.AssertExpectations(t)
	uow.Orders.AssertExpectations(t)
	uow.Carts.AssertExpectations(t)
	uow.Outbox.AssertExpectations(t)
}

func TestCheckout_EmptyOrMissingCart(t *testing.T) {
	t.Run("no active cart", func(t *testing.T) {
		uow := mocks.NewUnitOfWork()
		uow.Carts.On("LockByBuyer", mock.Anything, buyer, domain.StatusActive).Return(nil, domain.ErrCartNotFound)
		_, err := checkout.NewUsecase(uow).Checkout(context.Background(), buyer)
		assertAppError(t, err, http.StatusConflict)
	})
	t.Run("cart without items", func(t *testing.T) {
		uow := setup(domain.Cart{ID: 10, Items: []domain.Item{}}, nil)
		_, err := checkout.NewUsecase(uow).Checkout(context.Background(), buyer)
		assertAppError(t, err, http.StatusConflict)
		uow.Orders.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})
}

func TestCheckout_RejectsInvalidItems(t *testing.T) {
	tests := []struct {
		name   string
		modify func(ps []productdomain.Product) []productdomain.Product
		code   int
	}{
		{"product inactive", func(ps []productdomain.Product) []productdomain.Product {
			ps[1].Status = productdomain.StatusInactive
			return ps
		}, http.StatusConflict},
		{"product deleted", func(ps []productdomain.Product) []productdomain.Product { return ps[:1] }, http.StatusConflict},
		{"not enough stock", func(ps []productdomain.Product) []productdomain.Product {
			ps[0].Quantity = 1
			return ps
		}, http.StatusConflict},
		{"buyer owns the product", func(ps []productdomain.Product) []productdomain.Product {
			ps[0].SellerID = buyer
			return ps
		}, http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uow := setup(cart(), tt.modify(products()))

			_, err := checkout.NewUsecase(uow).Checkout(context.Background(), buyer)

			assertAppError(t, err, tt.code)
			assert.False(t, uow.Committed, "nothing may be committed")
			uow.Products.AssertNotCalled(t, "UpdateStock", mock.Anything, mock.Anything, mock.Anything)
			uow.Orders.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		})
	}
}

func TestCheckout_PriceChangedRefreshesSnapshotsAndPlacesNoOrder(t *testing.T) {
	ps := products()
	ps[0].DiscountPercentage = 10 // sale price is now 179.91
	uow := setup(cart(), ps)
	uow.Carts.On("UpdateItem", mock.Anything, mock.MatchedBy(func(i domain.Item) bool {
		return i.ID == 100 && i.PriceSnapshot == money.FromCents(17991) && i.Quantity == 2
	})).Return(nil)

	_, err := checkout.NewUsecase(uow).Checkout(context.Background(), buyer)

	assertAppError(t, err, http.StatusConflict)
	assert.True(t, uow.Committed, "refreshed snapshots must be committed")
	uow.Carts.AssertExpectations(t)
	uow.Products.AssertNotCalled(t, "UpdateStock", mock.Anything, mock.Anything, mock.Anything)
	uow.Orders.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	uow.Carts.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
}

func TestCheckout_CartAlreadyOrdered(t *testing.T) {
	uow := setup(cart(), products())
	uow.Products.On("UpdateStock", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	uow.Orders.On("Create", mock.Anything, mock.Anything).Return(nil, orderdomain.ErrCartAlreadyOrdered)

	_, err := checkout.NewUsecase(uow).Checkout(context.Background(), buyer)

	assertAppError(t, err, http.StatusConflict)
	assert.False(t, uow.Committed)
}

func TestCheckout_RepositoryFailureRollsBack(t *testing.T) {
	uow := setup(cart(), products())
	uow.Products.On("UpdateStock", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	uow.Orders.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("connection reset"))

	_, err := checkout.NewUsecase(uow).Checkout(context.Background(), buyer)

	assertAppError(t, err, http.StatusInternalServerError)
	assert.False(t, uow.Committed)
	uow.Carts.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
}

func TestCheckout_FailingToWriteAnEventRollsTheOrderBack(t *testing.T) {
	// Neither the order, nor the stock change, nor the closed cart may be
	// committed without the audit events.
	uow := setup(cart(), products())
	uow.Products.On("UpdateStock", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	uow.Orders.On("Create", mock.Anything, mock.Anything).Return(orderdomain.Order{ID: 55, TotalPrice: money.FromCents(38484)}, nil)
	uow.Outbox.On("Insert", mock.Anything, mock.Anything).Return(errors.New("outbox: insert event: connection reset"))

	_, err := checkout.NewUsecase(uow).Checkout(context.Background(), buyer)

	assertAppError(t, err, http.StatusInternalServerError)
	assert.False(t, uow.Committed)
	uow.Carts.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
}

func TestCheckout_NoEventsWhenNothingIsPlaced(t *testing.T) {
	// A price change refreshes the snapshots but places no order: nothing to audit.
	ps := products()
	ps[0].DiscountPercentage = 10
	uow := setup(cart(), ps)
	uow.Carts.On("UpdateItem", mock.Anything, mock.Anything).Return(nil)

	_, err := checkout.NewUsecase(uow).Checkout(context.Background(), buyer)

	assertAppError(t, err, http.StatusConflict)
	uow.Outbox.AssertNotCalled(t, "Insert", mock.Anything, mock.Anything)
}
