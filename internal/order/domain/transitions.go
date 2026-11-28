package domain

import "fmt"

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
