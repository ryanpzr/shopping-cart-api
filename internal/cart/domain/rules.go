package domain

import "fmt"

// ValidationError reports input that breaks a cart rule. Usecases map it to 400.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

// ValidateQuantity: an item quantity must be positive; removing an item is a
// separate operation (BR-CART-004).
func ValidateQuantity(quantity int) error {
	if quantity <= 0 {
		return &ValidationError{Field: "quantity", Message: "quantity must be greater than 0"}
	}
	return nil
}

// StockError reports that a product does not have enough units.
type StockError struct {
	ProductID int
	Title     string
	Requested int
	Available int
}

func (e *StockError) Error() string {
	return fmt.Sprintf("not enough stock for %q: requested %d, available %d", e.Title, e.Requested, e.Available)
}

// CheckStock fails when requested exceeds available.
func CheckStock(productID int, title string, requested, available int) error {
	if requested > available {
		return &StockError{ProductID: productID, Title: title, Requested: requested, Available: available}
	}
	return nil
}
