package domain

import (
	"errors"
	"time"

	"github.com/ryanpzr/shopping-cart-api/pkg/money"
)

const (
	StatusPending   = "pending"
	StatusPaid      = "paid"
	StatusShipped   = "shipped"
	StatusDelivered = "delivered"
	StatusCancelled = "cancelled"
)

// ErrCartAlreadyOrdered: a cart produces at most one order.
var ErrCartAlreadyOrdered = errors.New("cart already has an order")

type Order struct {
	ID         int
	CartID     int
	BuyerID    int
	TotalPrice money.Money
	Status     string
	Items      []Item
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Item keeps the price as it was at checkout: the unit price before the
// discount and the discount percentage.
type Item struct {
	ID                 int
	OrderID            int
	ProductID          int
	Quantity           int
	UnitPrice          money.Money // unit_price_snapshot
	DiscountPercentage int         // discount_snapshot
}

// SalePrice is the unit price after the discount.
func (i Item) SalePrice() money.Money {
	return i.UnitPrice.ApplyDiscount(i.DiscountPercentage)
}

// Subtotal is the sale price times the quantity.
func (i Item) Subtotal() money.Money {
	return i.SalePrice().Mul(i.Quantity)
}

// NewPending builds a pending order whose total is the sum of the item subtotals.
func NewPending(cartID, buyerID int, items []Item) Order {
	var total money.Money
	for _, item := range items {
		total += item.Subtotal()
	}
	return Order{
		CartID:     cartID,
		BuyerID:    buyerID,
		TotalPrice: total,
		Status:     StatusPending,
		Items:      items,
	}
}
