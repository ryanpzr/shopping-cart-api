package updateproduct

import (
	"time"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
)

// UpdateRequest is a partial update: omitted fields keep their current value.
// seller_id and status are not accepted here.
type UpdateRequest struct {
	Photo              *string  `json:"photo"`
	Title              *string  `json:"title"`
	Description        *string  `json:"description"`
	Price              *float64 `json:"price"`
	DiscountPercentage *int     `json:"discount_percentage"`
	Quantity           *int     `json:"quantity"`
	Category           *string  `json:"category"`
}

func (r UpdateRequest) isEmpty() bool {
	return r.Photo == nil && r.Title == nil && r.Description == nil && r.Price == nil &&
		r.DiscountPercentage == nil && r.Quantity == nil && r.Category == nil
}

type ProductResponse struct {
	ID                 int       `json:"id"`
	SellerID           int       `json:"seller_id"`
	Photo              *string   `json:"photo"`
	Title              string    `json:"title"`
	Description        *string   `json:"description"`
	Price              float64   `json:"price"`
	DiscountPercentage int       `json:"discount_percentage"`
	FinalPrice         float64   `json:"final_price"`
	Quantity           int       `json:"quantity"`
	Status             string    `json:"status"`
	Category           *string   `json:"category"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func toResponse(p domain.Product) ProductResponse {
	return ProductResponse{
		ID:                 p.ID,
		SellerID:           p.SellerID,
		Photo:              p.Photo,
		Title:              p.Title,
		Description:        p.Description,
		Price:              p.Price,
		DiscountPercentage: p.DiscountPercentage,
		FinalPrice:         p.FinalPrice(),
		Quantity:           p.Quantity,
		Status:             p.Status,
		Category:           p.Category,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
}
