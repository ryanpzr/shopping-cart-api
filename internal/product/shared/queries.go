package shared

const (
	queryFindById = `
		SELECT id, seller_id, photo, title, description, price, discount_percentage,
		       quantity, status, category, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	queryCreate = `
		INSERT INTO products (seller_id, photo, title, description, price, discount_percentage, quantity, status, category)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, seller_id, photo, title, description, price, discount_percentage,
		          quantity, status, category, created_at, updated_at
	`

	queryUpdate = `
		UPDATE products
		SET photo               = COALESCE($1, photo),
		    title               = COALESCE($2, title),
		    description         = COALESCE($3, description),
		    price               = COALESCE($4, price),
		    discount_percentage = COALESCE($5, discount_percentage),
		    quantity            = COALESCE($6, quantity),
		    category            = COALESCE($7, category),
		    updated_at          = NOW()
		WHERE id = $8
		RETURNING id, seller_id, photo, title, description, price, discount_percentage,
		          quantity, status, category, created_at, updated_at
	`

	queryUpdateStatus = `
		UPDATE products
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	queryDelete = `
		DELETE FROM products WHERE id = $1
	`

	// TODO(module-05): Esta query depende da tabela 'orders' criada em sdd/specs/modules/05-order.md.
	// Ao rodar DELETE antes de implementar o módulo 05, esta query retornará erro de relação inexistente.
	// Implementar o módulo 05 resolverá isso automaticamente.
	queryHasActiveOrders = `
		SELECT EXISTS(
			SELECT 1 FROM orders
			WHERE product_id = $1
			  AND status IN ('pending', 'paid', 'shipped')
		)
	`
)
