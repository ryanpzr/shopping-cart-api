package cmd

import (
	"log"
	"net/http"

	"github.com/ryanpzr/shopping-cart-api/internal/router"
)

func main() {
	r := router.SetupRouter()
	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Could not start server: %v\n", err)
	}
}
