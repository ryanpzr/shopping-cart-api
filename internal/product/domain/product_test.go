package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
)

func TestFinalPrice(t *testing.T) {
	tests := []struct {
		name     string
		price    float64
		discount int
		want     float64
	}{
		{"no discount", 100, 0, 100},
		{"ten percent", 100, 10, 90},
		{"full discount", 59.90, 100, 0},
		{"rounds to cents", 19.99, 15, 16.99}, // 16.9915
		{"rounds half up", 10.05, 50, 5.03},   // 5.025
		{"one third", 10, 33, 6.7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := domain.Product{Price: tt.price, DiscountPercentage: tt.discount}
			assert.InDelta(t, tt.want, p.FinalPrice(), 1e-9)
		})
	}
}

func TestIsActive(t *testing.T) {
	assert.True(t, domain.Product{Status: domain.StatusActive}.IsActive())
	assert.False(t, domain.Product{Status: domain.StatusInactive}.IsActive())
}

func assertInvalid(t *testing.T, err error) {
	t.Helper()
	var invalid *domain.ValidationError
	require.ErrorAs(t, err, &invalid)
	assert.NotEmpty(t, invalid.Field)
}

func TestValidateTitle(t *testing.T) {
	require.NoError(t, domain.ValidateTitle("Mechanical keyboard 75%"))
	require.NoError(t, domain.ValidateTitle(strings.Repeat("a", domain.MaxTitleLength)))
	assertInvalid(t, domain.ValidateTitle(""))
	assertInvalid(t, domain.ValidateTitle("   "))
	assertInvalid(t, domain.ValidateTitle(strings.Repeat("a", domain.MaxTitleLength+1)))
}

func TestValidatePrice(t *testing.T) {
	require.NoError(t, domain.ValidatePrice(0.01))
	require.NoError(t, domain.ValidatePrice(domain.MaxPrice))
	assertInvalid(t, domain.ValidatePrice(0))
	assertInvalid(t, domain.ValidatePrice(-1))
	assertInvalid(t, domain.ValidatePrice(domain.MaxPrice+0.01))
}

func TestValidateDiscount(t *testing.T) {
	require.NoError(t, domain.ValidateDiscount(0))
	require.NoError(t, domain.ValidateDiscount(80))
	assertInvalid(t, domain.ValidateDiscount(-1))
	assertInvalid(t, domain.ValidateDiscount(81))
}

func TestValidateQuantity(t *testing.T) {
	require.NoError(t, domain.ValidateQuantity(0))
	require.NoError(t, domain.ValidateQuantity(10))
	assertInvalid(t, domain.ValidateQuantity(-1))
}

func TestValidateCategory(t *testing.T) {
	require.NoError(t, domain.ValidateCategory(nil))
	for _, c := range domain.Categories {
		require.NoError(t, domain.ValidateCategory(&c), c)
	}
	for _, c := range []string{"books", "BOOKS", " Books", "Weapons", ""} {
		assertInvalid(t, domain.ValidateCategory(&c))
	}
}

func TestValidate(t *testing.T) {
	valid := domain.Product{Title: "Desk", Price: 100, DiscountPercentage: 10, Quantity: 0}
	require.NoError(t, valid.Validate())

	books := "Books"
	valid.Category = &books
	require.NoError(t, valid.Validate())

	tests := map[string]func(p *domain.Product){
		"title":               func(p *domain.Product) { p.Title = "" },
		"price":               func(p *domain.Product) { p.Price = 0 },
		"discount_percentage": func(p *domain.Product) { p.DiscountPercentage = 90 },
		"quantity":            func(p *domain.Product) { p.Quantity = -1 },
		"category":            func(p *domain.Product) { c := "books"; p.Category = &c },
	}
	for field, breakRule := range tests {
		t.Run(field, func(t *testing.T) {
			p := valid
			breakRule(&p)
			var invalid *domain.ValidationError
			require.ErrorAs(t, p.Validate(), &invalid)
			assert.Equal(t, field, invalid.Field)
		})
	}
}
