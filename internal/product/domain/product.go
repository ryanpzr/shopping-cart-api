package domain

import "time"

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
