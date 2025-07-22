package service

import (
	"github.com/ryanpzr/shopping-cart-api/internal/model"
	"github.com/ryanpzr/shopping-cart-api/internal/repository"
)

type Product interface {
	ChangeInfoProduct(id int, product model.Product) error
	GetAllProduct() ([]model.Product, error)
	PostNewProduct(product model.Product) (model.Product, error)
}

type productService struct {
	repo repository.Repository
}

func NewProductService(repo repository.Repository) Product {
	return &productService{repo: repo}
}

func (s *productService) ChangeInfoProduct(id int, product model.Product) error {
	return s.repo.ChangeInfoProduct(id, product)
}

func (s *productService) GetAllProduct() ([]model.Product, error) {
	return s.repo.GetAllProduct()
}

func (s *productService) PostNewProduct(product model.Product) (model.Product, error) {
	return s.repo.PostNewProduct(product)
}
