// Package shared is the MongoDB audit store of the activity log.
package shared

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/ryanpzr/shopping-cart-api/internal/activitylog/domain"
)

// CollectionName is the MongoDB collection that holds the audit entries.
const CollectionName = "activity_logs"

// Repository only talks to MongoDB. It returns domain errors or wrapped
// errors and holds no business rule.
type Repository interface {
	// Upsert stores the entry once: writing an event_id that already exists
	// is a no-op, so redeliveries of the same Kafka message are harmless.
	Upsert(ctx context.Context, entry domain.Entry) error
	// FindByUser returns one page of the user's entries, newest first, with
	// an optional event type filter, and the total number of matches. It
	// returns domain.ErrStoreUnavailable when MongoDB cannot be reached.
	FindByUser(ctx context.Context, userID int, eventType string, limit, offset int) ([]domain.Entry, int64, error)
	// EnsureIndexes creates the indexes of the collection. It is idempotent.
	EnsureIndexes(ctx context.Context) error
}

// document is the MongoDB representation of an entry.
type document struct {
	EventID    string         `bson:"event_id"`
	UserID     int            `bson:"user_id"`
	EventType  string         `bson:"event_type"`
	Metadata   map[string]any `bson:"metadata"`
	OccurredAt bson.DateTime  `bson:"occurred_at"`
	IngestedAt bson.DateTime  `bson:"ingested_at"`
}

type repository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &repository{collection: db.Collection(CollectionName)}
}

func wrap(op string, err error) error {
	return fmt.Errorf("activity log repository: %s: %w", op, err)
}

func (r *repository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		// Idempotency: one entry per event.
		{
			Keys:    bson.D{{Key: "event_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("ux_event_id"),
		},
		// BR-ADMIN-006 listing.
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "occurred_at", Value: -1}},
			Options: options.Index().SetName("ix_user_occurred_at"),
		},
		// Listing filtered by event_type.
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "event_type", Value: 1},
				{Key: "occurred_at", Value: -1},
			},
			Options: options.Index().SetName("ix_user_type_occurred_at"),
		},
	})
	if err != nil {
		return wrap("ensure indexes", err)
	}
	return nil
}

func (r *repository) Upsert(ctx context.Context, entry domain.Entry) error {
	metadata := entry.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	doc := document{
		EventID:    entry.EventID,
		UserID:     entry.UserID,
		EventType:  entry.EventType,
		Metadata:   metadata,
		OccurredAt: bson.NewDateTimeFromTime(entry.OccurredAt),
		IngestedAt: bson.NewDateTimeFromTime(entry.IngestedAt),
	}

	// $setOnInsert never rewrites an entry that already exists.
	_, err := r.collection.UpdateOne(ctx,
		bson.D{{Key: "event_id", Value: entry.EventID}},
		bson.D{{Key: "$setOnInsert", Value: doc}},
		options.UpdateOne().SetUpsert(true),
	)
	// Two concurrent upserts of the same event can both try to insert: the
	// unique index makes the loser fail, and the event is stored all the same.
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		return wrap("upsert", err)
	}
	return nil
}

func (r *repository) FindByUser(ctx context.Context, userID int, eventType string, limit, offset int) ([]domain.Entry, int64, error) {
	filter := bson.D{{Key: "user_id", Value: userID}}
	if eventType != "" {
		filter = append(filter, bson.E{Key: "event_type", Value: eventType})
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, unavailable("count", err)
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().
		SetSort(bson.D{{Key: "occurred_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit)))
	if err != nil {
		return nil, 0, unavailable("find", err)
	}
	defer func() { _ = cursor.Close(ctx) }()

	var docs []document
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, unavailable("read", err)
	}

	entries := make([]domain.Entry, 0, len(docs))
	for _, d := range docs {
		entries = append(entries, domain.Entry{
			EventID:    d.EventID,
			UserID:     d.UserID,
			EventType:  d.EventType,
			Metadata:   d.Metadata,
			OccurredAt: d.OccurredAt.Time().UTC(),
			IngestedAt: d.IngestedAt.Time().UTC(),
		})
	}
	return entries, total, nil
}

// unavailable reports a failure of the read path as the store being
// unavailable: the endpoint answers 503 and the cause stays in the log.
func unavailable(op string, err error) error {
	return fmt.Errorf("%w: %w", domain.ErrStoreUnavailable, wrap(op, err))
}
