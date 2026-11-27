package shared

const (
	orderColumns = `id, cart_id, buyer_id, total_price, status, created_at, updated_at`

	queryCreateOrder = `
		INSERT INTO orders (cart_id, buyer_id, total_price, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	queryCreateItem = `
		INSERT INTO order_items (order_id, product_id, quantity, unit_price_snapshot, discount_snapshot)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	queryFindByID = `SELECT ` + orderColumns + ` FROM orders WHERE id = $1`

	// Locks the order row, so concurrent cancellations and status changes of
	// the same order run one after the other.
	queryLockByID = queryFindByID + ` FOR UPDATE`

	queryCountByBuyer = `SELECT COUNT(*) FROM orders WHERE buyer_id = $1`

	queryListByBuyer = `
		SELECT ` + orderColumns + `
		FROM orders
		WHERE buyer_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`

	// Items of several orders in one query (no N+1). The title is for display.
	queryItemsByOrders = `
		SELECT oi.id, oi.order_id, oi.product_id, p.title, oi.quantity, oi.unit_price_snapshot, oi.discount_snapshot
		FROM order_items oi
		JOIN products p ON p.id = oi.product_id
		WHERE oi.order_id = ANY($1)
		ORDER BY oi.order_id, oi.id
	`

	queryUpdateStatus = `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	queryHasOrdersWithStatus = `
		SELECT EXISTS (
			SELECT 1
			FROM order_items oi
			JOIN orders o ON o.id = oi.order_id
			WHERE oi.product_id = $1
			  AND o.status = ANY($2)
		)
	`

	// Name of the unique constraint that allows one order per cart.
	constraintOneOrderPerCart = "orders_cart_id_key"

	pgUniqueViolation = "23505"
)
