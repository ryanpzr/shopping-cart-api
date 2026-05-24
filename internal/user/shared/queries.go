package shared

const (
	queryFindByEmail = `
		SELECT id, name, email, password_hash, role, status, timeout_until, created_at, updated_at
		FROM users WHERE email = $1
	`
	queryCreate = `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, role, status, created_at, updated_at
	`
	queryFindByID = `
		SELECT id, name, email, password_hash, role, status, timeout_until, created_at, updated_at
		FROM users WHERE id = $1
	`
	queryFindAll = `
		SELECT id, name, email, password_hash, role, status, timeout_until, created_at, updated_at
		FROM users ORDER BY id ASC
		LIMIT $1 OFFSET $2
	`
	queryCountAll = `SELECT COUNT(*) FROM users`

	queryUpdateProfile = `
		UPDATE users SET name = $1, email = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, email, role, status, created_at, updated_at
	`
	queryUpdateStatus = `
		UPDATE users SET status = $1, timeout_until = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, email, role, status, timeout_until, created_at, updated_at
	`
)
