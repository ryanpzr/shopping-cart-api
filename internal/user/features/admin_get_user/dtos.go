package admingetuser

import "time"

type AdminUserResponse struct {
	ID           int        `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	TimeoutUntil *time.Time `json:"timeout_until"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
