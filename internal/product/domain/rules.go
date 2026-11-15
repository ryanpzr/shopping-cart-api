package domain

import (
	"fmt"
	"strings"

	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

// Field limits from sdd/specs/modules/03-product.md.
const (
	MaxPrice          = 1_000_000.00
	MaxDiscount       = 100
	MaxTitleLength    = 255
	MaxCategoryLength = 100
)

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
		return apperrors.NewBadRequest("discount_percentage must be between 0 and 100")
	}
	return nil
}

func ValidateQuantity(quantity int) error {
	if quantity < 0 {
		return apperrors.NewBadRequest("quantity must be greater than or equal to 0")
	}
	return nil
}

func ValidateCategory(category *string) error {
	if category != nil && len([]rune(*category)) > MaxCategoryLength {
		return apperrors.NewBadRequest(fmt.Sprintf("category must be at most %d characters", MaxCategoryLength))
	}
	return nil
}
