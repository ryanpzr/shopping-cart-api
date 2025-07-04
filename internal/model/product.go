package model

type Product struct {
	Id          int
	Photo       string
	Title       string
	Description string
	Price       float64
	Quantity    int
}

type ProductDTO struct {
	Photo       *string  `json:"photo"`
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	Quantity    *int     `json:"quantity"`
}
