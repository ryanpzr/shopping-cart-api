package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ryanpzr/shopping-cart-api/internal/order/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/money"
)

func TestNewPending(t *testing.T) {
	items := []domain.Item{
		{ProductID: 1, Quantity: 2, UnitPrice: money.FromCents(19990), DiscountPercentage: 15}, // 169.92 x 2
		{ProductID: 2, Quantity: 1, UnitPrice: money.FromCents(4500)},                          // 45.00
	}

	order := domain.NewPending(3, 7, items)

	assert.Equal(t, domain.StatusPending, order.Status)
	assert.Equal(t, 3, order.CartID)
	assert.Equal(t, 7, order.BuyerID)
	assert.Equal(t, "169.92", items[0].SalePrice().String())
	assert.Equal(t, "339.84", items[0].Subtotal().String())
	assert.Equal(t, "384.84", order.TotalPrice.String())
	assert.Len(t, order.Items, 2)
}
