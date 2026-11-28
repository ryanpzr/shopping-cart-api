package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/ryanpzr/shopping-cart-api/config/pg"
	activitylogshared "github.com/ryanpzr/shopping-cart-api/internal/activitylog/shared"
	"github.com/ryanpzr/shopping-cart-api/internal/app/setup"
	jwtutil "github.com/ryanpzr/shopping-cart-api/pkg/jwt"
	"github.com/ryanpzr/shopping-cart-api/pkg/outbox"
)

const (
	// outboxRetention is how long events stay in the outbox after Debezium has
	// captured them: a short window for debugging and replays.
	outboxRetention  = 7 * 24 * time.Hour
	outboxPurgeEvery = time.Hour
)

func main() {
	// .env is optional: in containers the variables come from the environment.
	_ = godotenv.Load()

	if err := jwtutil.ValidateSecret(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	r := gin.Default()
	api := r.Group("/api/v1")

	db := pg.NewDb()
	conn, err := db.OpenConn()
	if err != nil {
		panic(err)
	}

	// The purger lives as long as the process; the DELETE is idempotent, so
	// several API instances can run it at the same time.
	outbox.StartPurger(context.Background(), conn, outboxPurgeEvery, outboxRetention)

	setup.Setup(api, conn, openActivityLog())

	log.Println("Starting server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Could not start server: %v\n", err)
	}
}

// openActivityLog returns the read side of the audit store. The API does not
// need MongoDB to run: without MONGO_URI, or while MongoDB is down, only
// GET /admin/users/:id/activity is affected (it answers 503).
func openActivityLog() activitylogshared.Repository {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Println("MONGO_URI is not set: GET /admin/users/:id/activity will answer 503")
		return activitylogshared.NewUnavailableRepository()
	}

	database := os.Getenv("MONGO_DB")
	if database == "" {
		database = activitylogshared.DefaultDatabase
	}

	// Connect does not wait for the server: it connects in the background.
	client, err := mongo.Connect(options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(2 * time.Second))
	if err != nil {
		log.Printf("Invalid MONGO_URI, GET /admin/users/:id/activity will answer 503: %v", err)
		return activitylogshared.NewUnavailableRepository()
	}
	return activitylogshared.NewRepository(client.Database(database))
}
