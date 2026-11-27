package checkout

import (
	"time"

	orderdomain "github.com/ryanpzr/shopping-cart-api/internal/order/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/money"
)

type OrderItemResponse struct {
	ProductID          int         `json:"product_id"`
	Quantity           int         `json:"quantity"`
	UnitPrice          money.Money `json:"unit_price"`
	DiscountPercentage int         `json:"discount_percentage"`
	SalePrice          money.Money `json:"sale_price"`
	Subtotal           money.Money `json:"subtotal"`
}

// OrderResponse is the order summary returned by checkout (BR-CART-005).
type OrderResponse struct {
	ID         int                 `json:"id"`
	CartID     int                 `json:"cart_id"`
	Status     string              `json:"status"`
	TotalPrice money.Money         `json:"total_price"`
	Items      []OrderItemResponse `json:"items"`
	CreatedAt  time.Time           `json:"created_at"`
}

func toResponse(o orderdomain.Order) OrderResponse {
	items := make([]OrderItemResponse, 0, len(o.Items))
	for _, i := range o.Items {
		items = append(items, OrderItemResponse{
			ProductID:          i.ProductID,
			Quantity:           i.Quantity,
			UnitPrice:          i.UnitPrice,
			DiscountPercentage: i.DiscountPercentage,
			SalePrice:          i.SalePrice(),
			Subtotal:           i.Subtotal(),
		})
	}
	return OrderResponse{
		ID: o.ID, CartID: o.CartID, Status: o.Status, TotalPrice: o.TotalPrice, Items: items, CreatedAt: o.CreatedAt,
	}
}
