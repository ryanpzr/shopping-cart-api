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

// Statuses lists every valid order status.
var Statuses = []string{StatusPending, StatusPaid, StatusShipped, StatusDelivered, StatusCancelled}

// nextStatus is the only forward step from each status. Lifecycle:
//
//	pending → paid → shipped → delivered
//	pending → cancelled
//
// cancelled and delivered are terminal.
var nextStatus = map[string]string{
	StatusPending: StatusPaid,
	StatusPaid:    StatusShipped,
	StatusShipped: StatusDelivered,
}

// ValidateStatus accepts one of Statuses.
func ValidateStatus(status string) error {
	for _, s := range Statuses {
		if s == status {
			return nil
		}
	}
	return &ValidationError{Field: "status", Message: fmt.Sprintf("status must be one of: %v", Statuses)}
}

// CanTransition reports whether an order may move from one status to another:
// only to the next step, or to cancelled while still pending.
func CanTransition(from, to string) bool {
	if to == StatusCancelled {
		return from == StatusPending
	}
	next, ok := nextStatus[from]
	return ok && next == to
}

// CheckTransition is CanTransition as an error.
func CheckTransition(from, to string) error {
	if !CanTransition(from, to) {
		return &TransitionError{From: from, To: to}
	}
	return nil
}

// IsCancellable: only a pending order can be cancelled (BR-ORDER-003).
func (o Order) IsCancellable() bool { return CanTransition(o.Status, StatusCancelled) }

// RestockQuantities returns the new stock of each product after giving back
// the units of the order items. stock maps product id to the current quantity.
// It fails when an item's product is missing from stock.
func RestockQuantities(items []Item, stock map[int]int) (map[int]int, error) {
	restored := make(map[int]int, len(items))
	for _, item := range items {
		current, ok := restored[item.ProductID]
		if !ok {
			current, ok = stock[item.ProductID]
			if !ok {
				return nil, fmt.Errorf("product %d is missing while restoring the stock of the order", item.ProductID)
			}
		}
		restored[item.ProductID] = current + item.Quantity
	}
	return restored, nil
}
