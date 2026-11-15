package domain_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
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

func assertBadRequest(t *testing.T, err error) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
}

func TestValidateTitle(t *testing.T) {
	require.NoError(t, domain.ValidateTitle("Mechanical keyboard 75%"))
	require.NoError(t, domain.ValidateTitle(strings.Repeat("a", domain.MaxTitleLength)))
	assertBadRequest(t, domain.ValidateTitle(""))
	assertBadRequest(t, domain.ValidateTitle("   "))
	assertBadRequest(t, domain.ValidateTitle(strings.Repeat("a", domain.MaxTitleLength+1)))
}

func TestValidatePrice(t *testing.T) {
	require.NoError(t, domain.ValidatePrice(0.01))
	require.NoError(t, domain.ValidatePrice(domain.MaxPrice))
	assertBadRequest(t, domain.ValidatePrice(0))
	assertBadRequest(t, domain.ValidatePrice(-1))
	assertBadRequest(t, domain.ValidatePrice(domain.MaxPrice+0.01))
}

func TestValidateDiscount(t *testing.T) {
	require.NoError(t, domain.ValidateDiscount(0))
	require.NoError(t, domain.ValidateDiscount(80))
	assertBadRequest(t, domain.ValidateDiscount(-1))
	assertBadRequest(t, domain.ValidateDiscount(81))
}

func TestValidateQuantity(t *testing.T) {
	require.NoError(t, domain.ValidateQuantity(0))
	require.NoError(t, domain.ValidateQuantity(10))
	assertBadRequest(t, domain.ValidateQuantity(-1))
}

func TestValidateCategory(t *testing.T) {
	require.NoError(t, domain.ValidateCategory(nil))
	for _, c := range domain.Categories {
		require.NoError(t, domain.ValidateCategory(&c), c)
	}
	for _, c := range []string{"books", "BOOKS", " Books", "Weapons", ""} {
		assertBadRequest(t, domain.ValidateCategory(&c))
	}
}
