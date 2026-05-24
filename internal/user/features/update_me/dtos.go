package updateme

import "time"

type UpdateMeRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateMeResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
