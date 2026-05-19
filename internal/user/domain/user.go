package domain

import "time"

type User struct {
	ID           int
	Name         string
	Email        string
	PasswordHash string
	Role         string
	Status       string
	TimeoutUntil *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
