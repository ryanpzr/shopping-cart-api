package cancellation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/order/cancellation"
	"github.com/ryanpzr/shopping-cart-api/internal/order/domain"
	"github.com/ryanpzr/shopping-cart-api/internal/order/shared"
	ordermocks "github.com/ryanpzr/shopping-cart-api/internal/order/shared/mocks"
	productdomain "github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/outbox"
	outboxmocks "github.com/ryanpzr/shopping-cart-api/pkg/outbox/mocks"
)

func pendingOrder() domain.Order {
	return domain.Order{ID: 55, BuyerID: 7, Status: domain.StatusPending, Items: []domain.Item{
		{ProductID: 9, Quantity: 2}, {ProductID: 4, Quantity: 1},
	}}
}

// actor is the admin who cancels in these tests: the event belongs to the actor,
// not to the buyer of the order.
const actor = 99

func run(uow *ordermocks.UnitOfWork, order domain.Order) error {
	return uow.Do(context.Background(), func(r shared.TxRepositories) error {
		return cancellation.Cancel(context.Background(), r, order, actor)
	})
}

func TestCancel_RestoresStockAndMarksCancelled(t *testing.T) {
	uow := ordermocks.NewUnitOfWork()
	uow.Products.On("LockByIDs", mock.Anything, []int{4, 9}).Return([]productdomain.Product{
		{ID: 4, Quantity: 0}, {ID: 9, Quantity: 3},
	}, nil)
	uow.Products.On("UpdateStock", mock.Anything, 4, 1).Return(nil)
	uow.Products.On("UpdateStock", mock.Anything, 9, 5).Return(nil)
	uow.Orders.On("UpdateStatus", mock.Anything, 55, domain.StatusCancelled).Return(nil)
	// The event belongs to whoever cancelled, with the buyer in the metadata.
	uow.Outbox.On("Insert", mock.Anything, outboxmocks.Matches(outbox.EventOrderCancelled, actor,
		map[string]any{"order_id": 55, "buyer_id": 7})).Return(nil)

	err := run(uow, pendingOrder())

	require.NoError(t, err)
	assert.True(t, uow.Committed)
	uow.Products.AssertExpectations(t)
	uow.Orders.AssertExpectations(t)
	uow.Outbox.AssertExpectations(t)
}

func TestCancel_NotCancellableTouchesNothing(t *testing.T) {
	for _, status := range []string{domain.StatusPaid, domain.StatusShipped, domain.StatusDelivered, domain.StatusCancelled} {
		t.Run(status, func(t *testing.T) {
			uow := ordermocks.NewUnitOfWork()
			order := pendingOrder()
			order.Status = status

			err := run(uow, order)

			var transition *domain.TransitionError
			require.ErrorAs(t, err, &transition)
			assert.False(t, uow.Committed)
			uow.Products.AssertNotCalled(t, "LockByIDs", mock.Anything, mock.Anything)
			uow.Products.AssertNotCalled(t, "UpdateStock", mock.Anything, mock.Anything, mock.Anything)
			uow.Orders.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestCancel_ProductMissingFailsWithoutWriting(t *testing.T) {
	uow := ordermocks.NewUnitOfWork()
	uow.Products.On("LockByIDs", mock.Anything, []int{4, 9}).Return([]productdomain.Product{{ID: 9, Quantity: 3}}, nil)

	err := run(uow, pendingOrder())

	require.Error(t, err)
	assert.False(t, uow.Committed)
	uow.Products.AssertNotCalled(t, "UpdateStock", mock.Anything, mock.Anything, mock.Anything)
	uow.Orders.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
}

func TestCancel_RepositoryFailuresRollBack(t *testing.T) {
	t.Run("lock fails", func(t *testing.T) {
		uow := ordermocks.NewUnitOfWork()
		uow.Products.On("LockByIDs", mock.Anything, mock.Anything).Return(nil, errors.New("connection reset"))
		require.Error(t, run(uow, pendingOrder()))
		assert.False(t, uow.Committed)
	})
	t.Run("stock update fails", func(t *testing.T) {
		uow := ordermocks.NewUnitOfWork()
		uow.Products.On("LockByIDs", mock.Anything, mock.Anything).Return([]productdomain.Product{{ID: 4}, {ID: 9}}, nil)
		uow.Products.On("UpdateStock", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("deadlock detected"))
		require.Error(t, run(uow, pendingOrder()))
		assert.False(t, uow.Committed)
		uow.Orders.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("status update fails", func(t *testing.T) {
		uow := ordermocks.NewUnitOfWork()
		uow.Products.On("LockByIDs", mock.Anything, mock.Anything).Return([]productdomain.Product{{ID: 4}, {ID: 9}}, nil)
		uow.Products.On("UpdateStock", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		uow.Orders.On("UpdateStatus", mock.Anything, 55, domain.StatusCancelled).Return(errors.New("connection reset"))
		require.Error(t, run(uow, pendingOrder()))
		assert.False(t, uow.Committed)
	})
}

func TestCancel_NotCancellableEmitsNoEvent(t *testing.T) {
	uow := ordermocks.NewUnitOfWork()
	order := pendingOrder()
	order.Status = domain.StatusShipped

	require.Error(t, run(uow, order))

	uow.Outbox.AssertNotCalled(t, "Insert", mock.Anything, mock.Anything)
}

func TestCancel_FailingToWriteTheEventRollsTheCancellationBack(t *testing.T) {
	uow := ordermocks.NewUnitOfWork()
	uow.Products.On("LockByIDs", mock.Anything, mock.Anything).Return([]productdomain.Product{{ID: 4}, {ID: 9}}, nil)
	uow.Products.On("UpdateStock", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	uow.Orders.On("UpdateStatus", mock.Anything, 55, domain.StatusCancelled).Return(nil)
	uow.Outbox.On("Insert", mock.Anything, mock.Anything).Return(errors.New("outbox: insert event: connection reset"))

	require.Error(t, run(uow, pendingOrder()))

	assert.False(t, uow.Committed, "the stock and the status must not be committed without the event")
}
