package listproducts

import "time"

type ListRequest struct {
	Category *string  `form:"category"`
	MinPrice *float64 `form:"min_price"`
	MaxPrice *float64 `form:"max_price"`
	Search   *string  `form:"search"`
	Limit    int      `form:"limit,default=20"`
	Offset   int      `form:"offset,default=0"`
}

type ProductResponse struct {
	ID                 int        `json:"id"`
	SellerID           int        `json:"seller_id"`
	Photo              *string    `json:"photo"`
	Title              string     `json:"title"`
	Description        *string    `json:"description"`
	Price              float64    `json:"price"`
	DiscountPercentage int        `json:"discount_percentage"`
	FinalPrice         float64    `json:"final_price"`
	Quantity           int        `json:"quantity"`
	Status             string     `json:"status"`
	Category           *string    `json:"category"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type MetaResponse struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type ListResponse struct {
	Data []ProductResponse `json:"data"`
	Meta MetaResponse      `json:"meta"`
}
