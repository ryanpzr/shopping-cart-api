// Package money represents amounts of money as integer cents, so sums and
// multiplications are exact. It reads and writes PostgreSQL NUMERIC(10,2)
// without going through float64 and marshals to JSON as a decimal number
// (e.g. 89.9), keeping the API contract unchanged.
package money

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Money is an amount in cents.
type Money int64

// FromCents builds an amount from cents.
func FromCents(cents int64) Money { return Money(cents) }

// FromFloat converts a decimal amount, rounding half away from zero to cents.
// Use it only at the boundary with code that still holds float64 prices.
func FromFloat(v float64) Money { return Money(math.Round(v * 100)) }

// Cents returns the amount in cents.
func (m Money) Cents() int64 { return int64(m) }

// Mul multiplies the amount by an integer quantity.
func (m Money) Mul(quantity int) Money { return m * Money(quantity) }

// ApplyDiscount returns the amount minus percent%, rounded half up to cents.
// percent must be between 0 and 100 and the amount must not be negative.
func (m Money) ApplyDiscount(percent int) Money {
	return (m*Money(100-percent) + 50) / 100
}

// Float64 returns the amount as a decimal number (for display and JSON).
func (m Money) Float64() float64 { return float64(m) / 100 }

// String formats the amount with two decimals, e.g. "89.90".
func (m Money) String() string {
	sign := ""
	c := int64(m)
	if c < 0 {
		sign = "-"
		c = -c
	}
	return fmt.Sprintf("%s%d.%02d", sign, c/100, c%100)
}

// Parse reads a decimal string with at most two decimal places.
func Parse(s string) (Money, error) {
	s = strings.TrimSpace(s)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(strings.TrimPrefix(s, "-"), "+")

	whole, frac, _ := strings.Cut(s, ".")
	if whole == "" {
		whole = "0"
	}
	if len(frac) > 2 {
		return 0, fmt.Errorf("money: %q has more than two decimal places", s)
	}
	frac += strings.Repeat("0", 2-len(frac))

	w, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("money: invalid amount %q", s)
	}
	f, err := strconv.ParseInt(frac, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("money: invalid amount %q", s)
	}
	cents := w*100 + f
	if neg {
		cents = -cents
	}
	return Money(cents), nil
}

// Scan implements sql.Scanner for NUMERIC columns.
func (m *Money) Scan(src any) error {
	switch v := src.(type) {
	case []byte:
		parsed, err := Parse(string(v))
		*m = parsed
		return err
	case string:
		parsed, err := Parse(v)
		*m = parsed
		return err
	case int64:
		*m = Money(v * 100)
		return nil
	case float64:
		*m = FromFloat(v)
		return nil
	default:
		return fmt.Errorf("money: cannot scan %T", src)
	}
}

// Value implements driver.Valuer, sending the amount as an exact decimal string.
func (m Money) Value() (driver.Value, error) { return m.String(), nil }

// MarshalJSON writes the amount as a decimal number.
func (m Money) MarshalJSON() ([]byte, error) { return []byte(m.String()), nil }

// UnmarshalJSON reads a decimal number with at most two decimal places.
func (m *Money) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		return fmt.Errorf("money: expected a number, got a string")
	}
	var n json.Number
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("money: %w", err)
	}
	parsed, err := Parse(n.String())
	if err != nil {
		return err
	}
	*m = parsed
	return nil
}
