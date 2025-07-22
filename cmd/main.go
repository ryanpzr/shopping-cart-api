package main

import (
	"github.com/ryanpzr/shopping-cart-api/internal/handler"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/config/pg"
	"github.com/ryanpzr/shopping-cart-api/internal/repository"
	"github.com/ryanpzr/shopping-cart-api/internal/router"
	"github.com/ryanpzr/shopping-cart-api/internal/service"
)

func main() {
	r := gin.Default()
	api := r.Group("/api/v1")

	db := pg.NewDb()
	conn, err := db.OpenConn()
	if err != nil {
		panic(err)
	}

	rpCart := repository.NewCartRepository(conn)
	svCart := service.NewCartService(rpCart)
	hdCart := handler.NewCartHandler(svCart)
	client := router.NewClient(hdCart)
	client.ClientRouters(api)

	rpProduct := repository.NewRepository(conn)
	svProduct := service.NewProductService(rpProduct)
	hdProduct := handler.NewHandlerProduct(svProduct)
	admin := router.NewAdmin(hdProduct)
	admin.AdminRouters(api)

	log.Println("Starting server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Could not start server: %v\n", err)
	}
}
