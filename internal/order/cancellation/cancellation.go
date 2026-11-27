// Package cancellation holds the cancellation of an order, shared by the
// buyer's cancel and the admin status change so that giving the stock back
// lives in one place.
package cancellation

import (
	"context"
	"sort"

	"github.com/ryanpzr/shopping-cart-api/internal/order/domain"
	"github.com/ryanpzr/shopping-cart-api/internal/order/shared"
)

// Cancel moves the order to cancelled and gives the units of every item back
// to the stock. It must run inside a transaction, with order loaded through
// LockByID, so its status is current and a concurrent cancellation waits.
//
// It returns *domain.TransitionError when the order is not cancellable.
func Cancel(ctx context.Context, r shared.TxRepositories, order domain.Order) error {
	if err := domain.CheckTransition(order.Status, domain.StatusCancelled); err != nil {
		return err
	}

	// Products are locked by id, the same order the checkout uses.
	locked, err := r.Products.LockByIDs(ctx, order.ProductIDs())
	if err != nil {
		return err
	}
	stock := make(map[int]int, len(locked))
	for _, p := range locked {
		stock[p.ID] = p.Quantity
	}

	restored, err := domain.RestockQuantities(order.Items, stock)
	if err != nil {
		return err
	}
	productIDs := make([]int, 0, len(restored))
	for productID := range restored {
		productIDs = append(productIDs, productID)
	}
	sort.Ints(productIDs)
	for _, productID := range productIDs {
		if err := r.Products.UpdateStock(ctx, productID, restored[productID]); err != nil {
			return err
		}
	}

	return r.Orders.UpdateStatus(ctx, order.ID, domain.StatusCancelled)
}
