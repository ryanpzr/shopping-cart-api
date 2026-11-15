package domain

import (
	"fmt"
	"strings"

	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

// Field limits from sdd/specs/modules/03-product.md.
const (
	MaxPrice       = 1_000_000.00 // guards against input errors and abuse
	MaxDiscount    = 80           // the marketplace keeps a minimum margin
	MaxTitleLength = 255
)

// Categories is the fixed product catalog taxonomy (exact, case-sensitive match).
var Categories = []string{
	"Electronics",
	"Clothing",
	"Books",
	"Home & Garden",
	"Sports",
	"Beauty",
	"Toys",
	"Food & Beverages",
	"Health",
	"Automotive",
	"Music & Movies",
	"Office Supplies",
}

func ValidateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return apperrors.NewBadRequest("title is required")
	}
	if len([]rune(title)) > MaxTitleLength {
		return apperrors.NewBadRequest(fmt.Sprintf("title must be at most %d characters", MaxTitleLength))
	}
	return nil
}

func ValidatePrice(price float64) error {
	if price <= 0 {
		return apperrors.NewBadRequest("price must be greater than 0")
	}
	if price > MaxPrice {
		return apperrors.NewBadRequest("price must not exceed 1000000")
	}
	return nil
}

func ValidateDiscount(discount int) error {
	if discount < 0 || discount > MaxDiscount {
		return apperrors.NewBadRequest(fmt.Sprintf("discount_percentage must be between 0 and %d", MaxDiscount))
	}
	return nil
}

func ValidateQuantity(quantity int) error {
	if quantity < 0 {
		return apperrors.NewBadRequest("quantity must be greater than or equal to 0")
	}
	return nil
}

// ValidateCategory accepts nil (no category) or one of Categories.
func ValidateCategory(category *string) error {
	if category == nil {
		return nil
	}
	for _, c := range Categories {
		if *category == c {
			return nil
		}
	}
	return apperrors.NewBadRequest("category must be one of: " + strings.Join(Categories, ", "))
}
