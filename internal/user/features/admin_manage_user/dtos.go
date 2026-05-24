package adminmanageuser

import "time"

type TimeoutRequest struct {
	DurationHours int `json:"duration_hours"`
}

type ManageUserResponse struct {
	ID           int        `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	TimeoutUntil *time.Time `json:"timeout_until"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
