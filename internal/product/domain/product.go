package domain

import (
	"math"
	"time"
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

// FinalPrice is the price after the discount, rounded to cents.
func (p Product) FinalPrice() float64 {
	v := p.Price * (1 - float64(p.DiscountPercentage)/100)
	return math.Round(v*100) / 100
}

func (p Product) IsActive() bool {
	return p.Status == StatusActive
}
