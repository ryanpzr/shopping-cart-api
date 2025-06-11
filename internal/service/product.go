package service

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/model"
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
		return
	}

	ctx.JSON(http.StatusOK, products)
}

func (s *productService) PostNewProduct(ctx *gin.Context) {
	var productDTO model.ProductDTO

	if err := ctx.BindJSON(&productDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode the body of requisition"})
		return
	}

	regex := regexp.MustCompile(`^[A-Za-z]+$`)
	if !regex.MatchString(productDTO.Titulo) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Title there isn't numbers"})
		return
	}

	if productDTO.Titulo == "" || productDTO.Descricao == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Title or Description not be empty"})
		return
	}

	if productDTO.Preco <= 0 || productDTO.Quantidade <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Price or Quantity not be negative or `0`"})
		return
	}

	product := model.Product{
		Id:         productDTO.Id,
		Foto:       productDTO.Foto,
		Titulo:     productDTO.Titulo,
		Descricao:  productDTO.Descricao,
		Preco:      productDTO.Preco,
		Quantidade: productDTO.Quantidade,
	}

	response, err := s.repo.Post(product)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, response)
}
