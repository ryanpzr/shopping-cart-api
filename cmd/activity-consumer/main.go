// Command activity-consumer moves audit events from Kafka into MongoDB.
//
// It reads the activity-log.events topic (fed by Debezium from the outbox
// table) in the consumer group activity-log-writer, stores each event once and
// only then commits its offset. Messages that cannot be decoded go to the
// dead-letter topic.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/ryanpzr/shopping-cart-api/internal/activitylog/consumer"
	activitylogshared "github.com/ryanpzr/shopping-cart-api/internal/activitylog/shared"
)

func main() {
	_ = godotenv.Load()

	cfg, err := consumer.LoadConfig(os.Getenv)
	if err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client, err := mongo.Connect(options.Client().
		ApplyURI(cfg.MongoURI).
		SetServerSelectionTimeout(5 * time.Second))
	if err != nil {
		log.Fatalf("Could not create the MongoDB client: %v", err)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	repo := activitylogshared.NewRepository(client.Database(cfg.MongoDB))
	if err := ensureIndexes(ctx, repo); err != nil {
		log.Fatalf("Could not create the MongoDB indexes: %v", err)
	}

	reader := consumer.NewReader(cfg.Kafka)
	defer func() { _ = reader.Close() }()
	dlq := consumer.NewDLQPublisher(cfg.Kafka)
	defer func() { _ = dlq.Close() }()

	handler := consumer.NewHandler(repo, dlq, consumer.Backoff{Initial: 500 * time.Millisecond, Max: 30 * time.Second})

	log.Printf("activity consumer started: topic %s, group %s, brokers %v",
		cfg.Kafka.Topic, cfg.Kafka.GroupID, cfg.Kafka.Brokers)
	if err := consumer.Run(ctx, reader, handler); err != nil {
		log.Fatalf("Consumer stopped: %v", err)
	}
	log.Println("activity consumer stopped")
}

// ensureIndexes waits for MongoDB to accept connections, which can take a
// while right after the stack starts, and creates the indexes.
func ensureIndexes(ctx context.Context, repo activitylogshared.Repository) error {
	for {
		err := repo.EnsureIndexes(ctx)
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		log.Printf("MongoDB is not ready yet, retrying: %v", err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}
