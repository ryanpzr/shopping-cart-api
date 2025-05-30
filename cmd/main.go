package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/router"
	"github.com/ryanpzr/shopping-cart-api/internal/service"
)

func main() {
	r := gin.Default()
	api := r.Group("/api")

	cartService := service.NewCartService()
	client := router.NewClient(cartService)
	client.ClientRouters(api)

	productService := service.NewProductService()
	admin := router.NewAdmin(productService)
	admin.AdminRouters(api)

	log.Println("Starting server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Could not start server: %v\n", err)
	}
}
