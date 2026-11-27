package shared

const (
	cartColumns = `id, buyer_id, status, created_at, updated_at`

	queryFindByBuyer = `
		SELECT ` + cartColumns + `
		FROM carts
		WHERE buyer_id = $1 AND status = $2
	`

	queryLockByBuyer = queryFindByBuyer + ` FOR UPDATE`

	queryCreate = `
		INSERT INTO carts (buyer_id, status)
		VALUES ($1, $2)
		RETURNING ` + cartColumns

	queryUpdateStatus = `
		UPDATE carts
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	queryListItems = `
		SELECT ci.id, ci.cart_id, ci.product_id, p.title, ci.quantity, ci.price_snapshot, ci.created_at
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1
		ORDER BY ci.id
	`

	queryAddItem = `
		INSERT INTO cart_items (cart_id, product_id, quantity, price_snapshot)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	queryUpdateItem = `
		UPDATE cart_items
		SET quantity = $1, price_snapshot = $2
		WHERE id = $3 AND cart_id = $4
	`

	queryRemoveItem = `DELETE FROM cart_items WHERE id = $1 AND cart_id = $2`

	// Name of the partial unique index that allows one active cart per buyer.
	constraintOneActiveCart = "ux_carts_one_active_per_buyer"
)
