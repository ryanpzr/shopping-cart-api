package adminupdatestatus

import (
	"time"

	"github.com/ryanpzr/shopping-cart-api/internal/order/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/money"
)

type UpdateStatusRequest struct {
	Status *string `json:"status"`
}

type ItemResponse struct {
	ProductID          int         `json:"product_id"`
	Title              string      `json:"title"`
	Quantity           int         `json:"quantity"`
	UnitPrice          money.Money `json:"unit_price"`
	DiscountPercentage int         `json:"discount_percentage"`
	SalePrice          money.Money `json:"sale_price"`
	Subtotal           money.Money `json:"subtotal"`
}

type OrderResponse struct {
	ID         int            `json:"id"`
	CartID     int            `json:"cart_id"`
	Status     string         `json:"status"`
	TotalPrice money.Money    `json:"total_price"`
	Items      []ItemResponse `json:"items"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func toResponse(o domain.Order) OrderResponse {
	items := make([]ItemResponse, 0, len(o.Items))
	for _, i := range o.Items {
		items = append(items, ItemResponse{
			ProductID:          i.ProductID,
			Title:              i.Title,
			Quantity:           i.Quantity,
			UnitPrice:          i.UnitPrice,
			DiscountPercentage: i.DiscountPercentage,
			SalePrice:          i.SalePrice(),
			Subtotal:           i.Subtotal(),
		})
	}
	return OrderResponse{
		ID: o.ID, CartID: o.CartID, Status: o.Status, TotalPrice: o.TotalPrice,
		Items: items, CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt,
	}
}
