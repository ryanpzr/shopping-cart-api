package cancelorder_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/order/domain"
	cancelorder "github.com/ryanpzr/shopping-cart-api/internal/order/features/cancel_order"
	ordermocks "github.com/ryanpzr/shopping-cart-api/internal/order/shared/mocks"
	productdomain "github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
	"github.com/ryanpzr/shopping-cart-api/pkg/outbox"
	outboxmocks "github.com/ryanpzr/shopping-cart-api/pkg/outbox/mocks"
)

const (
	owner   = 7
	orderID = 55
)

func assertAppError(t *testing.T, err error, code int) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, code, appErr.Code)
}

func order(status string) domain.Order {
	return domain.Order{ID: orderID, BuyerID: owner, Status: status, Items: []domain.Item{
		{ProductID: 9, Title: "Desk", Quantity: 2}, {ProductID: 4, Title: "Lamp", Quantity: 1},
	}}
}

func TestCancel_OwnerCancelsPendingOrderAndStockReturns(t *testing.T) {
	uow := ordermocks.NewUnitOfWork()
	uow.Orders.On("LockByID", mock.Anything, orderID).Return(order(domain.StatusPending), nil)
	uow.Products.On("LockByIDs", mock.Anything, []int{4, 9}).Return([]productdomain.Product{{ID: 4, Quantity: 0}, {ID: 9, Quantity: 3}}, nil)
	uow.Products.On("UpdateStock", mock.Anything, 4, 1).Return(nil)
	uow.Products.On("UpdateStock", mock.Anything, 9, 5).Return(nil)
	uow.Orders.On("UpdateStatus", mock.Anything, orderID, domain.StatusCancelled).Return(nil)
	uow.Orders.On("FindByID", mock.Anything, orderID).Return(order(domain.StatusCancelled), nil)
	uow.Outbox.On("Insert", mock.Anything, outboxmocks.Matches(outbox.EventOrderCancelled, owner,
		map[string]any{"order_id": orderID, "buyer_id": owner})).Return(nil)

	resp, err := cancelorder.NewUsecase(uow).Cancel(context.Background(), orderID, owner)

	require.NoError(t, err)
	assert.Equal(t, domain.StatusCancelled, resp.Status, "the response is read in the same transaction")
	assert.True(t, uow.Committed)
	uow.Products.AssertExpectations(t)
	uow.Orders.AssertExpectations(t)
	uow.Outbox.AssertExpectations(t)
}

func TestCancel_OnlyPendingOrdersAreCancellable(t *testing.T) {
	for _, status := range []string{domain.StatusPaid, domain.StatusShipped, domain.StatusDelivered, domain.StatusCancelled} {
		t.Run(status, func(t *testing.T) {
			uow := ordermocks.NewUnitOfWork()
			uow.Orders.On("LockByID", mock.Anything, orderID).Return(order(status), nil)

			_, err := cancelorder.NewUsecase(uow).Cancel(context.Background(), orderID, owner)

			assertAppError(t, err, http.StatusConflict)
			assert.False(t, uow.Committed)
			uow.Products.AssertNotCalled(t, "UpdateStock", mock.Anything, mock.Anything, mock.Anything)
			uow.Outbox.AssertNotCalled(t, "Insert", mock.Anything, mock.Anything)
		})
	}
}

func TestCancel_AnotherUserIsForbiddenBeforeAnyChange(t *testing.T) {
	uow := ordermocks.NewUnitOfWork()
	uow.Orders.On("LockByID", mock.Anything, orderID).Return(order(domain.StatusPending), nil)

	_, err := cancelorder.NewUsecase(uow).Cancel(context.Background(), orderID, 8)

	assertAppError(t, err, http.StatusForbidden)
	assert.False(t, uow.Committed)
	uow.Products.AssertNotCalled(t, "LockByIDs", mock.Anything, mock.Anything)
}

func TestCancel_AnotherUsersNonPendingOrderIsStillForbidden(t *testing.T) {
	// Ownership is checked before the status, so the answer does not leak the status.
	uow := ordermocks.NewUnitOfWork()
	uow.Orders.On("LockByID", mock.Anything, orderID).Return(order(domain.StatusDelivered), nil)

	_, err := cancelorder.NewUsecase(uow).Cancel(context.Background(), orderID, 8)

	assertAppError(t, err, http.StatusForbidden)
}

func TestCancel_NotFound(t *testing.T) {
	uow := ordermocks.NewUnitOfWork()
	uow.Orders.On("LockByID", mock.Anything, orderID).Return(nil, domain.ErrOrderNotFound)

	_, err := cancelorder.NewUsecase(uow).Cancel(context.Background(), orderID, owner)

	assertAppError(t, err, http.StatusNotFound)
}

func TestCancel_FailureWhileRestoringRollsBack(t *testing.T) {
	uow := ordermocks.NewUnitOfWork()
	uow.Orders.On("LockByID", mock.Anything, orderID).Return(order(domain.StatusPending), nil)
	uow.Products.On("LockByIDs", mock.Anything, mock.Anything).Return([]productdomain.Product{{ID: 4}, {ID: 9}}, nil)
	uow.Products.On("UpdateStock", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("deadlock detected"))

	_, err := cancelorder.NewUsecase(uow).Cancel(context.Background(), orderID, owner)

	assertAppError(t, err, http.StatusInternalServerError)
	assert.False(t, uow.Committed)
	uow.Orders.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
}
