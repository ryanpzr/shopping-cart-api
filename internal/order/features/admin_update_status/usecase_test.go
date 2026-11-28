package adminupdatestatus_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/order/domain"
	adminupdatestatus "github.com/ryanpzr/shopping-cart-api/internal/order/features/admin_update_status"
	ordermocks "github.com/ryanpzr/shopping-cart-api/internal/order/shared/mocks"
	productdomain "github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
	"github.com/ryanpzr/shopping-cart-api/pkg/outbox"
	outboxmocks "github.com/ryanpzr/shopping-cart-api/pkg/outbox/mocks"
)

const (
	orderID = 55
	admin   = 1 // the admin acting on the order; the buyer is user 7
)

func ptr[T any](v T) *T { return &v }

func assertAppError(t *testing.T, err error, code int) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, code, appErr.Code)
}

func order(status string) domain.Order {
	return domain.Order{ID: orderID, BuyerID: 7, Status: status, Items: []domain.Item{{ProductID: 9, Title: "Desk", Quantity: 2}}}
}

func TestUpdateStatus_AllowedSteps(t *testing.T) {
	for _, step := range [][2]string{
		{domain.StatusPending, domain.StatusPaid},
		{domain.StatusPaid, domain.StatusShipped},
		{domain.StatusShipped, domain.StatusDelivered},
	} {
		t.Run(step[0]+" to "+step[1], func(t *testing.T) {
			uow := ordermocks.NewUnitOfWork()
			uow.Orders.On("LockByID", mock.Anything, orderID).Return(order(step[0]), nil)
			uow.Orders.On("UpdateStatus", mock.Anything, orderID, step[1]).Return(nil)
			uow.Orders.On("FindByID", mock.Anything, orderID).Return(order(step[1]), nil)

			resp, err := adminupdatestatus.NewUsecase(uow).UpdateStatus(context.Background(), orderID, admin,
				adminupdatestatus.UpdateStatusRequest{Status: ptr(step[1])})

			require.NoError(t, err)
			assert.Equal(t, step[1], resp.Status)
			assert.True(t, uow.Committed)
			// Moving forward never touches the products: no lock, no stock change.
			uow.Products.AssertNotCalled(t, "LockByIDs", mock.Anything, mock.Anything)
			uow.Outbox.AssertNotCalled(t, "Insert", mock.Anything, mock.Anything)
			uow.Products.AssertNotCalled(t, "UpdateStock", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestUpdateStatus_EveryOtherChangeIsAConflict(t *testing.T) {
	allowed := map[[2]string]bool{
		{domain.StatusPending, domain.StatusPaid}:      true,
		{domain.StatusPaid, domain.StatusShipped}:      true,
		{domain.StatusShipped, domain.StatusDelivered}: true,
		{domain.StatusPending, domain.StatusCancelled}: true,
	}
	for _, from := range domain.Statuses {
		for _, to := range domain.Statuses {
			if allowed[[2]string{from, to}] {
				continue
			}
			t.Run(from+" to "+to, func(t *testing.T) {
				uow := ordermocks.NewUnitOfWork()
				uow.Orders.On("LockByID", mock.Anything, orderID).Return(order(from), nil)

				_, err := adminupdatestatus.NewUsecase(uow).UpdateStatus(context.Background(), orderID, admin,
					adminupdatestatus.UpdateStatusRequest{Status: ptr(to)})

				assertAppError(t, err, http.StatusConflict)
				assert.False(t, uow.Committed)
				uow.Orders.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
				uow.Products.AssertNotCalled(t, "UpdateStock", mock.Anything, mock.Anything, mock.Anything)
			})
		}
	}
}

func TestUpdateStatus_AdminCancelRestoresStock(t *testing.T) {
	uow := ordermocks.NewUnitOfWork()
	uow.Orders.On("LockByID", mock.Anything, orderID).Return(order(domain.StatusPending), nil)
	uow.Products.On("LockByIDs", mock.Anything, []int{9}).Return([]productdomain.Product{{ID: 9, Quantity: 1}}, nil)
	uow.Products.On("UpdateStock", mock.Anything, 9, 3).Return(nil)
	uow.Orders.On("UpdateStatus", mock.Anything, orderID, domain.StatusCancelled).Return(nil)
	uow.Orders.On("FindByID", mock.Anything, orderID).Return(order(domain.StatusCancelled), nil)
	// The event belongs to the admin who cancelled, with the buyer in the metadata.
	uow.Outbox.On("Insert", mock.Anything, outboxmocks.Matches(outbox.EventOrderCancelled, admin,
		map[string]any{"order_id": orderID, "buyer_id": 7})).Return(nil)

	resp, err := adminupdatestatus.NewUsecase(uow).UpdateStatus(context.Background(), orderID, admin,
		adminupdatestatus.UpdateStatusRequest{Status: ptr(domain.StatusCancelled)})

	require.NoError(t, err)
	assert.Equal(t, domain.StatusCancelled, resp.Status)
	assert.True(t, uow.Committed)
	uow.Products.AssertExpectations(t)
	uow.Outbox.AssertExpectations(t)
}

func TestUpdateStatus_CancelOfANonPendingOrderDoesNotTouchTheStock(t *testing.T) {
	uow := ordermocks.NewUnitOfWork()
	uow.Orders.On("LockByID", mock.Anything, orderID).Return(order(domain.StatusShipped), nil)

	_, err := adminupdatestatus.NewUsecase(uow).UpdateStatus(context.Background(), orderID, admin,
		adminupdatestatus.UpdateStatusRequest{Status: ptr(domain.StatusCancelled)})

	assertAppError(t, err, http.StatusConflict)
	uow.Products.AssertNotCalled(t, "UpdateStock", mock.Anything, mock.Anything, mock.Anything)
}

func TestUpdateStatus_BadRequests(t *testing.T) {
	for name, status := range map[string]*string{
		"missing":   nil,
		"empty":     ptr(""),
		"unknown":   ptr("refunded"),
		"wrongcase": ptr("PAID"),
	} {
		t.Run(name, func(t *testing.T) {
			uow := ordermocks.NewUnitOfWork()

			_, err := adminupdatestatus.NewUsecase(uow).UpdateStatus(context.Background(), orderID, admin,
				adminupdatestatus.UpdateStatusRequest{Status: status})

			assertAppError(t, err, http.StatusBadRequest)
			uow.Orders.AssertNotCalled(t, "LockByID", mock.Anything, mock.Anything)
		})
	}
}

func TestUpdateStatus_NotFound(t *testing.T) {
	uow := ordermocks.NewUnitOfWork()
	uow.Orders.On("LockByID", mock.Anything, orderID).Return(nil, domain.ErrOrderNotFound)

	_, err := adminupdatestatus.NewUsecase(uow).UpdateStatus(context.Background(), orderID, admin,
		adminupdatestatus.UpdateStatusRequest{Status: ptr(domain.StatusPaid)})

	assertAppError(t, err, http.StatusNotFound)
}
