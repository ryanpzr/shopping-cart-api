package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/repository"
)

type Product interface {
	ChangeInfoProduct(ctx *gin.Context)
	GetAllProduct(ctx *gin.Context)
	PostNewProduct(ctx *gin.Context)
}

type productService struct {
	repo repository.Repository
}

func NewProductService(repo repository.Repository) Product {
	return &productService{repo: repo}
}

func (s *productService) ChangeInfoProduct(ctx *gin.Context) {
}

func (s *productService) GetAllProduct(ctx *gin.Context) {
	products, err := s.repo.GetAllProduct()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get products"})
	}

	ctx.JSON(http.StatusOK, products)
}

func (s *productService) PostNewProduct(ctx *gin.Context) {
}
