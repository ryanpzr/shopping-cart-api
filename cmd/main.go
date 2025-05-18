package cmd

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/router"
)

func main() {
	r := gin.Default()

	api := r.Group("/api")
	router.ClientRouters(api)
	router.AdminRouters(api)

	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Could not start server: %v\n", err)
	}
}
