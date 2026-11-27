package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/order/domain"
)

func TestCanTransition_FullMatrix(t *testing.T) {
	allowed := map[[2]string]bool{
		{domain.StatusPending, domain.StatusPaid}:      true,
		{domain.StatusPaid, domain.StatusShipped}:      true,
		{domain.StatusShipped, domain.StatusDelivered}: true,
		{domain.StatusPending, domain.StatusCancelled}: true,
	}
	for _, from := range domain.Statuses {
		for _, to := range domain.Statuses {
			want := allowed[[2]string{from, to}]
			assert.Equal(t, want, domain.CanTransition(from, to), "%s -> %s", from, to)
		}
	}
}

func TestCheckTransition(t *testing.T) {
	require.NoError(t, domain.CheckTransition(domain.StatusPending, domain.StatusPaid))

	var transition *domain.TransitionError
	require.ErrorAs(t, domain.CheckTransition(domain.StatusCancelled, domain.StatusPending), &transition)
	assert.Equal(t, domain.StatusCancelled, transition.From)
	assert.Equal(t, domain.StatusPending, transition.To)
	assert.Contains(t, transition.Error(), "cancelled")
}

func TestUnknownStatusesNeverTransition(t *testing.T) {
	assert.False(t, domain.CanTransition("refunded", domain.StatusPaid))
	assert.False(t, domain.CanTransition(domain.StatusPending, "refunded"))
}

func TestIsCancellable(t *testing.T) {
	assert.True(t, domain.Order{Status: domain.StatusPending}.IsCancellable())
	for _, s := range []string{domain.StatusPaid, domain.StatusShipped, domain.StatusDelivered, domain.StatusCancelled} {
		assert.False(t, domain.Order{Status: s}.IsCancellable(), s)
	}
}

func TestValidateStatus(t *testing.T) {
	for _, s := range domain.Statuses {
		require.NoError(t, domain.ValidateStatus(s), s)
	}
	for _, s := range []string{"", "PAID", "refunded", " paid"} {
		var invalid *domain.ValidationError
		require.ErrorAs(t, domain.ValidateStatus(s), &invalid, s)
		assert.Equal(t, "status", invalid.Field)
	}
}

func TestRestockQuantities(t *testing.T) {
	items := []domain.Item{{ProductID: 1, Quantity: 2}, {ProductID: 2, Quantity: 1}}

	restored, err := domain.RestockQuantities(items, map[int]int{1: 3, 2: 0, 9: 50})

	require.NoError(t, err)
	assert.Equal(t, map[int]int{1: 5, 2: 1}, restored, "only the order's products, current stock plus the items")
}

func TestRestockQuantities_SumsRepeatedProducts(t *testing.T) {
	items := []domain.Item{{ProductID: 1, Quantity: 2}, {ProductID: 1, Quantity: 3}}

	restored, err := domain.RestockQuantities(items, map[int]int{1: 10})

	require.NoError(t, err)
	assert.Equal(t, map[int]int{1: 15}, restored)
}

func TestRestockQuantities_MissingProduct(t *testing.T) {
	_, err := domain.RestockQuantities([]domain.Item{{ProductID: 7, Quantity: 1}}, map[int]int{})
	assert.Error(t, err)
}

func TestProductIDsAreSorted(t *testing.T) {
	o := domain.Order{Items: []domain.Item{{ProductID: 9}, {ProductID: 2}, {ProductID: 5}}}
	assert.Equal(t, []int{2, 5, 9}, o.ProductIDs())
}

func TestOwnedBy(t *testing.T) {
	assert.True(t, domain.Order{BuyerID: 7}.OwnedBy(7))
	assert.False(t, domain.Order{BuyerID: 7}.OwnedBy(8))
}
