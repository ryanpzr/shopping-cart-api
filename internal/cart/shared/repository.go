package shared

import (
	"database/sql"
)

type Repository interface {
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

type repository struct {
	db *sql.DB
}
