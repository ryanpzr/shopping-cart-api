package domain

import "errors"

// ErrProductNotFound is returned by the repository when no product matches.
// Usecases decide how to present it (e.g. 404).
var ErrProductNotFound = errors.New("product not found")

// ErrProductInUse is returned by the repository when other records (order
// items) still reference the product, so it cannot be deleted.
var ErrProductInUse = errors.New("product is referenced by other records")

// ValidationError reports a product field that breaks a business rule.
// It carries no transport concerns: usecases map it to a 400.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

func invalid(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
