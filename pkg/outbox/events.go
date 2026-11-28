// Package outbox implements the producer side of the transactional outbox used
// by the activity log: events are inserted into outbox_events in the same
// transaction as the business change they describe, and Debezium publishes
// them to Kafka (see sdd/specs/modules/06-activity-log.md).
package outbox

// Event types of the audit trail (Event Catalog in the spec).
const (
	EventLogin          = "login"
	EventProductCreated = "product_created"
	EventProductUpdated = "product_updated"
	EventProductDeleted = "product_deleted"
	EventCartCheckout   = "cart_checkout"
	EventOrderPlaced    = "order_placed"
	EventOrderCancelled = "order_cancelled"
)

// EventTypes lists every valid event type.
var EventTypes = []string{
	EventLogin,
	EventProductCreated,
	EventProductUpdated,
	EventProductDeleted,
	EventCartCheckout,
	EventOrderPlaced,
	EventOrderCancelled,
}

// aggregates maps an event type to the kind of entity it is about.
var aggregates = map[string]string{
	EventLogin:          "user",
	EventProductCreated: "product",
	EventProductUpdated: "product",
	EventProductDeleted: "product",
	EventCartCheckout:   "cart",
	EventOrderPlaced:    "order",
	EventOrderCancelled: "order",
}

// ValidType reports whether t is a known event type.
func ValidType(t string) bool {
	_, ok := aggregates[t]
	return ok
}

// AggregateType returns the kind of entity the event type is about.
func AggregateType(eventType string) string {
	return aggregates[eventType]
}
