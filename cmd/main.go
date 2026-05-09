package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/ryanpzr/shopping-cart-api/config/pg"
	"github.com/ryanpzr/shopping-cart-api/internal/app/setup"
)

func main() {
	godotenv.Load()

	r := gin.Default()
	api := r.Group("/api/v1")

	db := pg.NewDb()
	conn, err := db.OpenConn()
	if err != nil {
		panic(err)
	}

	setup.Setup(api, conn)

	log.Println("Starting server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Could not start server: %v\n", err)
	}
}
