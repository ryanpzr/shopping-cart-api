package domain

import (
	"errors"
	"time"

	"github.com/ryanpzr/shopping-cart-api/pkg/money"
)

const (
	StatusActive     = "active"
	StatusCheckedOut = "checked_out"
)

var (
	// ErrCartNotFound: the buyer has no cart with the requested status.
	ErrCartNotFound = errors.New("cart not found")
	// ErrItemNotFound: the item is not in the cart.
	ErrItemNotFound = errors.New("cart item not found")
	// ErrActiveCartExists: creating an active cart lost a race with another request.
	ErrActiveCartExists = errors.New("buyer already has an active cart")
)

type Cart struct {
	ID        int
	BuyerID   int
	Status    string
	Items     []Item
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Item struct {
	ID            int
	CartID        int
	ProductID     int
	Title         string // product title, for display
	Quantity      int
	PriceSnapshot money.Money // sale price when the item was added or last refreshed
	CreatedAt     time.Time
}

// Subtotal is the item price times its quantity.
func (i Item) Subtotal() money.Money {
	return i.PriceSnapshot.Mul(i.Quantity)
}

// Total is the sum of the item subtotals. Discounts are already part of each
// price snapshot.
func (c Cart) Total() money.Money {
	var total money.Money
	for _, item := range c.Items {
		total += item.Subtotal()
	}
	return total
}

// ItemByID returns the item with the given id, if it is in the cart.
func (c Cart) ItemByID(id int) (Item, bool) {
	for _, item := range c.Items {
		if item.ID == id {
			return item, true
		}
	}
	return Item{}, false
}

// ItemByProduct returns the item for the given product, if it is in the cart.
func (c Cart) ItemByProduct(productID int) (Item, bool) {
	for _, item := range c.Items {
		if item.ProductID == productID {
			return item, true
		}
	}
	return Item{}, false
}

// ProductIDs lists the products in the cart.
func (c Cart) ProductIDs() []int {
	ids := make([]int, 0, len(c.Items))
	for _, item := range c.Items {
		ids = append(ids, item.ProductID)
	}
	return ids
}

func (c Cart) IsEmpty() bool { return len(c.Items) == 0 }
