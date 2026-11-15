package shared

const productColumns = `id, seller_id, photo, title, description, price, discount_percentage,
       quantity, status, category, created_at, updated_at`

const (
	queryFindByID = `
		SELECT ` + productColumns + `
		FROM products
		WHERE id = $1
	`

	queryCreate = `
		INSERT INTO products (seller_id, photo, title, description, price, discount_percentage, quantity, status, category)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING ` + productColumns

	// Writes every editable column. seller_id never changes and status has its
	// own operation (queryUpdateStatus).
	queryUpdate = `
		UPDATE products
		SET photo               = $1,
		    title               = $2,
		    description         = $3,
		    price               = $4,
		    discount_percentage = $5,
		    quantity            = $6,
		    category            = $7,
		    updated_at          = NOW()
		WHERE id = $8
		RETURNING ` + productColumns

	queryUpdateStatus = `
		UPDATE products
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	queryDelete = `DELETE FROM products WHERE id = $1`

	queryHasOrdersWithStatus = `
		SELECT EXISTS (
			SELECT 1
			FROM order_items oi
			JOIN orders o ON o.id = oi.order_id
			WHERE oi.product_id = $1
			  AND o.status = ANY($2)
		)
	`
)
