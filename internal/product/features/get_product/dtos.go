package getproduct

import "time"

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
