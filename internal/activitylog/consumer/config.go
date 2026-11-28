package consumer

import (
	"errors"
	"strings"

	activitylogshared "github.com/ryanpzr/shopping-cart-api/internal/activitylog/shared"
)

// Config is the whole configuration of the activity consumer.
type Config struct {
	Kafka    KafkaConfig
	MongoURI string
	MongoDB  string
}

// LoadConfig reads the configuration through getenv (os.Getenv in production).
// MONGO_URI is required; everything else has a default that matches the
// docker-compose profile.
func LoadConfig(getenv func(string) string) (Config, error) {
	cfg := Config{
		Kafka: KafkaConfig{
			Brokers:  splitList(valueOr(getenv("KAFKA_BROKERS"), "localhost:9092")),
			Topic:    valueOr(getenv("KAFKA_TOPIC"), "activity-log.events"),
			GroupID:  valueOr(getenv("KAFKA_GROUP"), "activity-log-writer"),
			DLQTopic: valueOr(getenv("KAFKA_DLQ_TOPIC"), "activity-log.events.dlq"),
		},
		MongoURI: strings.TrimSpace(getenv("MONGO_URI")),
		MongoDB:  valueOr(getenv("MONGO_DB"), activitylogshared.DefaultDatabase),
	}
	if cfg.MongoURI == "" {
		return Config{}, errors.New("MONGO_URI is required")
	}
	if len(cfg.Kafka.Brokers) == 0 {
		return Config{}, errors.New("KAFKA_BROKERS has no broker")
	}
	return cfg, nil
}

func valueOr(v, fallback string) string {
	if v = strings.TrimSpace(v); v != "" {
		return v
	}
	return fallback
}

func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
