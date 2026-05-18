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
)
