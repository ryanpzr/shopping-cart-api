package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/money"
)

func cart() domain.Cart {
	return domain.Cart{ID: 1, Items: []domain.Item{
		{ID: 10, ProductID: 5, Quantity: 2, PriceSnapshot: money.FromCents(8990)},
		{ID: 11, ProductID: 7, Quantity: 3, PriceSnapshot: money.FromCents(3333)},
	}}
}

func TestTotalIsExact(t *testing.T) {
	c := cart()
	assert.Equal(t, "179.80", c.Items[0].Subtotal().String())
	assert.Equal(t, "99.99", c.Items[1].Subtotal().String())
	assert.Equal(t, "279.79", c.Total().String())
	assert.Equal(t, "0.00", domain.Cart{}.Total().String())
}

func TestLookups(t *testing.T) {
	c := cart()
	item, ok := c.ItemByID(11)
	require.True(t, ok)
	assert.Equal(t, 7, item.ProductID)
	_, ok = c.ItemByID(99)
	assert.False(t, ok)

	item, ok = c.ItemByProduct(5)
	require.True(t, ok)
	assert.Equal(t, 10, item.ID)
	_, ok = c.ItemByProduct(99)
	assert.False(t, ok)

	assert.Equal(t, []int{5, 7}, c.ProductIDs())
	assert.False(t, c.IsEmpty())
	assert.True(t, domain.Cart{}.IsEmpty())
}

func TestValidateQuantity(t *testing.T) {
	require.NoError(t, domain.ValidateQuantity(1))
	for _, q := range []int{0, -1} {
		var invalid *domain.ValidationError
		require.ErrorAs(t, domain.ValidateQuantity(q), &invalid)
		assert.Equal(t, "quantity", invalid.Field)
	}
}

func TestCheckStock(t *testing.T) {
	require.NoError(t, domain.CheckStock(5, "Lamp", 3, 3))
	var stock *domain.StockError
	require.ErrorAs(t, domain.CheckStock(5, "Lamp", 4, 3), &stock)
	assert.Equal(t, `not enough stock for "Lamp": requested 4, available 3`, stock.Error())
}
