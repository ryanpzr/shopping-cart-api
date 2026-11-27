package additemtocart

import (
	"github.com/ryanpzr/shopping-cart-api/internal/cart/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/money"
)

type AddItemRequest struct {
	ProductID *int `json:"product_id"`
	Quantity  *int `json:"quantity"`
}

type ItemResponse struct {
	ID            int         `json:"id"`
	ProductID     int         `json:"product_id"`
	Title         string      `json:"title"`
	Quantity      int         `json:"quantity"`
	PriceSnapshot money.Money `json:"price_snapshot"`
}

type CartResponse struct {
	ID            int            `json:"id"`
	Status        string         `json:"status"`
	Items         []ItemResponse `json:"items"`
	Subtotal      money.Money    `json:"subtotal"`
	DiscountTotal money.Money    `json:"discount_total"`
	Total         money.Money    `json:"total"`
}

func toResponse(c domain.Cart) CartResponse {
	items := make([]ItemResponse, 0, len(c.Items))
	for _, i := range c.Items {
		items = append(items, ItemResponse{
			ID: i.ID, ProductID: i.ProductID, Title: i.Title, Quantity: i.Quantity, PriceSnapshot: i.PriceSnapshot,
		})
	}
	total := c.Total()
	return CartResponse{ID: c.ID, Status: c.Status, Items: items, Subtotal: total, Total: total}
}
