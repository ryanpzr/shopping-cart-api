package checkout

import (
	"context"
	"errors"
	"fmt"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/domain"
	orderdomain "github.com/ryanpzr/shopping-cart-api/internal/order/domain"
	productdomain "github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

// errPricesChanged: some prices changed since the items were added. The
// snapshots are refreshed and committed, so the buyer can review the cart and
// check out again.
var errPricesChanged = apperrors.NewConflict(
	"prices changed since the items were added; the cart was updated, review it and check out again")

type usecase struct {
	uow UnitOfWork
}

func NewUsecase(uow UnitOfWork) Usecase {
	return &usecase{uow: uow}
}

// Checkout implements BR-CART-005 in a single transaction: lock the cart and
// its products, re-validate every item, decrement the stock, create the order
// and close the cart. Locks make concurrent checkouts wait for each other, so
// stock cannot be oversold.
func (u *usecase) Checkout(ctx context.Context, buyerID int) (OrderResponse, error) {
	var (
		order         orderdomain.Order
		pricesChanged bool
	)

	err := u.uow.Do(ctx, func(r TxRepositories) error {
		cart, err := r.Carts.LockByBuyer(ctx, buyerID, domain.StatusActive)
		if errors.Is(err, domain.ErrCartNotFound) || (err == nil && cart.IsEmpty()) {
			return apperrors.NewConflict("cart is empty")
		}
		if err != nil {
			return err
		}

		locked, err := r.Products.LockByIDs(ctx, cart.ProductIDs())
		if err != nil {
			return err
		}
		products := make(map[int]productdomain.Product, len(locked))
		for _, p := range locked {
			products[p.ID] = p
		}

		for _, item := range cart.Items {
			product, ok := products[item.ProductID]
			if !ok || !product.IsActive() {
				return apperrors.NewConflict(fmt.Sprintf("%q is no longer available", item.Title))
			}
			if product.SellerID == buyerID {
				return apperrors.NewForbidden("you cannot buy your own product")
			}
			if err := domain.CheckStock(product.ID, product.Title, item.Quantity, product.Quantity); err != nil {
				return err
			}
			if current := product.SalePrice(); current != item.PriceSnapshot {
				item.PriceSnapshot = current
				if err := r.Carts.UpdateItem(ctx, item); err != nil {
					return err
				}
				pricesChanged = true
			}
		}
		if pricesChanged {
			return nil // commit the refreshed snapshots, place no order
		}

		items := make([]orderdomain.Item, 0, len(cart.Items))
		for _, item := range cart.Items {
			product := products[item.ProductID]
			if err := r.Products.UpdateStock(ctx, product.ID, product.Quantity-item.Quantity); err != nil {
				return err
			}
			items = append(items, orderdomain.Item{
				ProductID:          product.ID,
				Quantity:           item.Quantity,
				UnitPrice:          product.UnitPrice(),
				DiscountPercentage: product.DiscountPercentage,
			})
		}

		order, err = r.Orders.Create(ctx, orderdomain.NewPending(cart.ID, buyerID, items))
		if err != nil {
			return err
		}

		// TODO(module-06): emit `cart_checkout` and `order_placed` through the outbox, in this transaction.

		return r.Carts.UpdateStatus(ctx, cart.ID, domain.StatusCheckedOut)
	})
	if err != nil {
		return OrderResponse{}, toAppError(err)
	}
	if pricesChanged {
		return OrderResponse{}, errPricesChanged
	}
	return toResponse(order), nil
}

func toAppError(err error) error {
	var (
		appErr *apperrors.AppError
		stock  *domain.StockError
	)
	switch {
	case errors.As(err, &appErr):
		return appErr
	case errors.As(err, &stock):
		return apperrors.NewConflict(stock.Error())
	default:
		return apperrors.NewInternalServer(err.Error())
	}
}
