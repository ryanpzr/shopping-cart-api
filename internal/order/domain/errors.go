package domain

import (
	"errors"
	"fmt"
)

// ErrOrderNotFound is returned by the repository when no order matches.
// Usecases decide how to present it (e.g. 404).
var ErrOrderNotFound = errors.New("order not found")

// ValidationError reports input that breaks an order rule. Usecases map it to 400.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

// TransitionError reports a status change the order lifecycle does not allow.
// Usecases map it to 409.
type TransitionError struct {
	From string
	To   string
}

func (e *TransitionError) Error() string {
	return fmt.Sprintf("an order in status %q cannot change to %q", e.From, e.To)
}
