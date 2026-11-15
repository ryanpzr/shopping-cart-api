package domain

import "errors"

// ErrProductNotFound is returned by the repository when no product matches.
// Usecases decide how to present it (e.g. 404).
var ErrProductNotFound = errors.New("product not found")

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
