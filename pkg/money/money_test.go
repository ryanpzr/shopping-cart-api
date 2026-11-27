package money_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/pkg/money"
)

func TestParse(t *testing.T) {
	tests := map[string]int64{
		"89.90": 8990, "89.9": 8990, "89": 8900, "0.01": 1, ".5": 50, "-1.25": -125, " 10.00 ": 1000,
	}
	for in, want := range tests {
		got, err := money.Parse(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got.Cents(), in)
	}
	for _, bad := range []string{"1.234", "abc", "1.x"} {
		_, err := money.Parse(bad)
		assert.Error(t, err, bad)
	}
}

func TestString(t *testing.T) {
	assert.Equal(t, "89.90", money.FromCents(8990).String())
	assert.Equal(t, "0.05", money.FromCents(5).String())
	assert.Equal(t, "-1.25", money.FromCents(-125).String())
}

func TestArithmeticIsExact(t *testing.T) {
	// 0.1 + 0.2 in float64 is 0.30000000000000004; in cents it is exact.
	sum := money.FromCents(10) + money.FromCents(20)
	assert.Equal(t, "0.30", sum.String())

	// 3 x 33.33 = 99.99, and summing 1000 lines of 0.10 gives exactly 100.00.
	assert.Equal(t, "99.99", money.FromCents(3333).Mul(3).String())
	var total money.Money
	for i := 0; i < 1000; i++ {
		total += money.FromCents(10)
	}
	assert.Equal(t, "100.00", total.String())
}

func TestApplyDiscount(t *testing.T) {
	tests := []struct {
		cents    int64
		percent  int
		expected string
	}{
		{10000, 0, "100.00"},
		{10000, 10, "90.00"},
		{19990, 15, "169.92"}, // 169.915 rounds up
		{1005, 50, "5.03"},    // 5.025 rounds up
		{1000, 33, "6.70"},
		{5990, 100, "0.00"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, money.FromCents(tt.cents).ApplyDiscount(tt.percent).String())
	}
}

func TestFromFloat(t *testing.T) {
	assert.Equal(t, int64(16992), money.FromFloat(169.915).Cents()) // rounds half up
	assert.Equal(t, int64(8990), money.FromFloat(89.9).Cents())
	assert.Equal(t, int64(30), money.FromFloat(0.1+0.2).Cents())
}

func TestScan(t *testing.T) {
	var m money.Money
	require.NoError(t, m.Scan([]byte("199.90")))
	assert.Equal(t, int64(19990), m.Cents())
	require.NoError(t, m.Scan("5.5"))
	assert.Equal(t, int64(550), m.Cents())
	require.NoError(t, m.Scan(int64(7)))
	assert.Equal(t, int64(700), m.Cents())
	assert.Error(t, m.Scan(true))
}

func TestValue(t *testing.T) {
	v, err := money.FromCents(8990).Value()
	require.NoError(t, err)
	assert.Equal(t, "89.90", v)
}

func TestJSON(t *testing.T) {
	b, err := json.Marshal(struct {
		Total money.Money `json:"total"`
	}{money.FromCents(17980)})
	require.NoError(t, err)
	assert.JSONEq(t, `{"total":179.80}`, string(b))

	var in struct {
		Price money.Money `json:"price"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{"price":89.9}`), &in))
	assert.Equal(t, int64(8990), in.Price.Cents())
	assert.Error(t, json.Unmarshal([]byte(`{"price":"89.9"}`), &in))
	assert.Error(t, json.Unmarshal([]byte(`{"price":1.999}`), &in))
}
