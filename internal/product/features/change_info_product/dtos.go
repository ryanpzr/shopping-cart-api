package changeinfoproduct

type ProductDTO struct {
	Photo       *string  `json:"photo"`
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	Quantity    *int     `json:"quantity"`
}
