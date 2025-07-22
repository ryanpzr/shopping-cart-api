package repository

import "database/sql"

func NewCartRepository(db *sql.DB) Cart {
	return &cart{db: db}
}

type Cart interface {
}

type cart struct {
	db *sql.DB
}
