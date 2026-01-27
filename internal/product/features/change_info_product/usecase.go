package changeinfoproduct

import (
	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
)

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (s *usecase) ChangeInfoProduct(id int, product domain.Product) error {
	return s.repo.ChangeInfoProduct(id, product)
}

func (s *usecase) GetAllProduct() ([]domain.Product, error) {
	return s.repo.GetAllProduct()
}

func (s *usecase) PostNewProduct(product domain.Product) (domain.Product, error) {
	return s.repo.PostNewProduct(product)
}
