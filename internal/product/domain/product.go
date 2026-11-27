package domain

import (
	"time"

	"github.com/ryanpzr/shopping-cart-api/pkg/money"
)

const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

type Product struct {
	ID                 int
	SellerID           int
	Photo              *string
	Title              string
	Description        *string
	Price              float64
	DiscountPercentage int
	Quantity           int
	Status             string
	Category           *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// UnitPrice is the price before the discount, as money.
func (p Product) UnitPrice() money.Money {
	return money.FromFloat(p.Price)
}

// SalePrice is the price after the discount, rounded half up to cents. It is
// the single source of the price a buyer pays (cart snapshots, orders).
func (p Product) SalePrice() money.Money {
	return p.UnitPrice().ApplyDiscount(p.DiscountPercentage)
}

// FinalPrice is SalePrice as a decimal number, for the product API responses.
func (p Product) FinalPrice() float64 {
	return p.SalePrice().Float64()
}

func (p Product) IsActive() bool {
	return p.Status == StatusActive
}
