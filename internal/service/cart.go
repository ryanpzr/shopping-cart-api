package service

import (
	"github.com/ryanpzr/shopping-cart-api/internal/repository"
)

type Cart interface {
}

type cartService struct {
	rp repository.Cart
}

func NewCartService(rp repository.Cart) Cart {
	return &cartService{
		rp: rp,
	}
}
