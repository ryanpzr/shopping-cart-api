package consumer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/activitylog/domain"
)

const validEvent = `{"event_id":"0b6f1c2e-1111-4222-8333-444455556666","user_id":42,"event_type":"cart_checkout",` +
	`"metadata":{"order_id":10,"total_price":179.8},"occurred_at":"2026-11-28T10:30:00Z"}`

// fakeStore records the entries it receives and fails with the queued errors first.
type fakeStore struct {
	entries []domain.Entry
	errs    []error
}

func (s *fakeStore) Upsert(_ context.Context, e domain.Entry) error {
	if len(s.errs) > 0 {
		err := s.errs[0]
		s.errs = s.errs[1:]
		if err != nil {
			return err
		}
	}
	s.entries = append(s.entries, e)
	return nil
}

type dlqRecord struct {
	key, value []byte
	reason     string
}

type fakeDLQ struct {
	records []dlqRecord
	err     error
}

func (d *fakeDLQ) Publish(_ context.Context, key, value []byte, reason string) error {
	if d.err != nil {
		return d.err
	}
	d.records = append(d.records, dlqRecord{key, value, reason})
	return nil
}

func newHandler(store *fakeStore, dlq *fakeDLQ) *Handler {
	h := NewHandler(store, dlq, Backoff{Initial: time.Second, Max: 8 * time.Second})
	h.now = func() time.Time { return time.Date(2026, 11, 28, 10, 30, 1, 0, time.UTC) }
	h.wait = func(context.Context, time.Duration) error { return nil }
	return h
}

func TestHandle_StoresTheEvent(t *testing.T) {
	store, dlq := &fakeStore{}, &fakeDLQ{}

	err := newHandler(store, dlq).Handle(context.Background(), []byte("42"), []byte(validEvent))

	require.NoError(t, err)
	require.Len(t, store.entries, 1)
	e := store.entries[0]
	assert.Equal(t, "0b6f1c2e-1111-4222-8333-444455556666", e.EventID)
	assert.Equal(t, 42, e.UserID)
	assert.Equal(t, "cart_checkout", e.EventType)
	assert.EqualValues(t, 10, e.Metadata["order_id"])
	assert.Equal(t, time.Date(2026, 11, 28, 10, 30, 0, 0, time.UTC), e.OccurredAt)
	assert.Equal(t, time.Date(2026, 11, 28, 10, 30, 1, 0, time.UTC), e.IngestedAt)
	assert.Empty(t, dlq.records)
}

func TestHandle_RedeliveryReachesTheStoreWithTheSameEventID(t *testing.T) {
	// The unique event_id index (repository.Upsert, covered by the smoke test) absorbs
	// repeats; here we only check that a redelivery is not dead-lettered and carries the same
	// event_id to the store.
	store, dlq := &fakeStore{}, &fakeDLQ{}
	h := newHandler(store, dlq)

	require.NoError(t, h.Handle(context.Background(), []byte("42"), []byte(validEvent)))
	require.NoError(t, h.Handle(context.Background(), []byte("42"), []byte(validEvent)))

	assert.Len(t, store.entries, 2, "the consumer asks to store both; the unique event_id index absorbs the repeat")
	assert.Equal(t, store.entries[0].EventID, store.entries[1].EventID)
	assert.Empty(t, dlq.records)
}

func TestHandle_MalformedMessagesGoToTheDeadLetters(t *testing.T) {
	tests := map[string]string{
		"not json":               `this is not json`,
		"empty":                  ``,
		"json array":             `[1,2,3]`,
		"no event id":            `{"user_id":1,"event_type":"login","metadata":{},"occurred_at":"2026-11-28T10:30:00Z"}`,
		"bad event id":           `{"event_id":"abc","user_id":1,"event_type":"login","metadata":{},"occurred_at":"2026-11-28T10:30:00Z"}`,
		"no user":                `{"event_id":"0b6f1c2e-1111-4222-8333-444455556666","event_type":"login","metadata":{},"occurred_at":"2026-11-28T10:30:00Z"}`,
		"unknown type":           `{"event_id":"0b6f1c2e-1111-4222-8333-444455556666","user_id":1,"event_type":"user_banned","metadata":{},"occurred_at":"2026-11-28T10:30:00Z"}`,
		"no time":                `{"event_id":"0b6f1c2e-1111-4222-8333-444455556666","user_id":1,"event_type":"login","metadata":{}}`,
		"time of the wrong type": `{"event_id":"0b6f1c2e-1111-4222-8333-444455556666","user_id":1,"event_type":"login","metadata":{},"occurred_at":5}`,
	}
	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			store, dlq := &fakeStore{}, &fakeDLQ{}

			err := newHandler(store, dlq).Handle(context.Background(), []byte("7"), []byte(value))

			require.NoError(t, err, "a poison message must not block the partition: the offset is committed")
			assert.Empty(t, store.entries)
			require.Len(t, dlq.records, 1)
			assert.Equal(t, []byte("7"), dlq.records[0].key)
			assert.Equal(t, []byte(value), dlq.records[0].value, "the original bytes are kept")
			assert.NotEmpty(t, dlq.records[0].reason)
		})
	}
}

func TestHandle_FailingToPublishToTheDeadLettersIsTransient(t *testing.T) {
	store, dlq := &fakeStore{}, &fakeDLQ{err: errors.New("kafka down")}

	err := newHandler(store, dlq).Handle(context.Background(), []byte("7"), []byte("garbage"))

	require.Error(t, err, "the offset must not be committed: the message would be lost")
}

func TestHandle_StoreErrorIsTransient(t *testing.T) {
	store, dlq := &fakeStore{errs: []error{errors.New("mongo unavailable")}}, &fakeDLQ{}

	err := newHandler(store, dlq).Handle(context.Background(), []byte("42"), []byte(validEvent))

	require.Error(t, err)
	assert.Empty(t, dlq.records, "a valid event is never dead-lettered because of a store failure")
}

func TestProcess_RetriesWithExponentialBackoffUntilTheStoreRecovers(t *testing.T) {
	boom := errors.New("mongo unavailable")
	store, dlq := &fakeStore{errs: []error{boom, boom, boom, boom, boom}}, &fakeDLQ{}
	h := newHandler(store, dlq)
	var waits []time.Duration
	h.wait = func(_ context.Context, d time.Duration) error { waits = append(waits, d); return nil }
	var attempts []int
	onRetry := func(attempt int, _ error) { attempts = append(attempts, attempt) }

	err := h.Process(context.Background(), []byte("42"), []byte(validEvent), onRetry)

	require.NoError(t, err)
	assert.Len(t, store.entries, 1)
	assert.Equal(t, []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 8 * time.Second}, waits,
		"doubles from the initial delay and stops at the maximum")
	assert.Equal(t, []int{0, 1, 2, 3, 4}, attempts)
}

func TestProcess_StopsWhenTheContextEnds(t *testing.T) {
	store, dlq := &fakeStore{errs: []error{errors.New("down"), errors.New("down"), errors.New("down")}}, &fakeDLQ{}
	h := newHandler(store, dlq)
	ctx, cancel := context.WithCancel(context.Background())
	h.wait = func(ctx context.Context, _ time.Duration) error { cancel(); return ctx.Err() }

	err := h.Process(ctx, []byte("42"), []byte(validEvent), nil)

	require.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, store.entries)
}

func TestBackoffDelay(t *testing.T) {
	b := Backoff{Initial: 200 * time.Millisecond, Max: time.Second}
	assert.Equal(t, 200*time.Millisecond, b.delay(0))
	assert.Equal(t, 400*time.Millisecond, b.delay(1))
	assert.Equal(t, 800*time.Millisecond, b.delay(2))
	assert.Equal(t, time.Second, b.delay(3))
	assert.Equal(t, time.Second, b.delay(40))
}

// fakeSource serves queued messages, records commits and cancels the context
// when the queue is empty, like a shutdown signal.
type fakeSource struct {
	queue     []kafka.Message
	committed []kafka.Message
	commitErr error
	fetchErr  error
	cancel    context.CancelFunc
	log       *[]string
}

func (s *fakeSource) FetchMessage(ctx context.Context) (kafka.Message, error) {
	if s.fetchErr != nil {
		return kafka.Message{}, s.fetchErr
	}
	if len(s.queue) == 0 {
		s.cancel()
		return kafka.Message{}, ctx.Err()
	}
	m := s.queue[0]
	s.queue = s.queue[1:]
	return m, nil
}

func (s *fakeSource) CommitMessages(_ context.Context, msgs ...kafka.Message) error {
	if s.commitErr != nil {
		return s.commitErr
	}
	s.committed = append(s.committed, msgs...)
	if s.log != nil {
		*s.log = append(*s.log, "commit")
	}
	return nil
}

func message(offset int64, value string) kafka.Message {
	return kafka.Message{Partition: 0, Offset: offset, Key: []byte("42"), Value: []byte(value)}
}

func TestRun_CommitsEachOffsetAfterStoringIt(t *testing.T) {
	var order []string
	store := &orderedStore{log: &order}
	ctx, cancel := context.WithCancel(context.Background())
	src := &fakeSource{queue: []kafka.Message{message(0, validEvent), message(1, validEvent)}, cancel: cancel, log: &order}
	h := NewHandler(store, &fakeDLQ{}, Backoff{Initial: time.Millisecond, Max: time.Millisecond})

	require.NoError(t, Run(ctx, src, h))

	assert.Equal(t, []string{"store", "commit", "store", "commit"}, order, "the offset is committed only after the entry is stored")
	assert.Len(t, src.committed, 2)
}

// orderedStore logs when it stores, to prove the order against the commits.
type orderedStore struct{ log *[]string }

func (s *orderedStore) Upsert(context.Context, domain.Entry) error {
	*s.log = append(*s.log, "store")
	return nil
}

func TestRun_AMalformedMessageDoesNotBlockTheOnesAfterIt(t *testing.T) {
	store, dlq := &fakeStore{}, &fakeDLQ{}
	ctx, cancel := context.WithCancel(context.Background())
	src := &fakeSource{
		queue:  []kafka.Message{message(0, "garbage"), message(1, validEvent)},
		cancel: cancel,
	}
	h := NewHandler(store, dlq, Backoff{Initial: time.Millisecond, Max: time.Millisecond})

	require.NoError(t, Run(ctx, src, h))

	assert.Len(t, dlq.records, 1)
	assert.Len(t, store.entries, 1)
	assert.Len(t, src.committed, 2, "both offsets are committed")
}

func TestRun_DoesNotCommitWhileTheStoreIsDown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	boom := errors.New("mongo unavailable")
	store := &fakeStore{errs: []error{boom, boom, boom}}
	src := &fakeSource{queue: []kafka.Message{message(0, validEvent)}, cancel: cancel}
	h := NewHandler(store, &fakeDLQ{}, Backoff{Initial: time.Millisecond, Max: time.Millisecond})
	h.wait = func(ctx context.Context, _ time.Duration) error {
		if len(store.errs) == 1 { // third failure: the process is told to stop
			cancel()
		}
		return ctx.Err()
	}

	require.NoError(t, Run(ctx, src, h))

	assert.Empty(t, src.committed, "an event that could not be stored must be redelivered")
	assert.Empty(t, store.entries)
}

func TestRun_ReturnsAnErrorWhenTheSourceOrTheCommitFails(t *testing.T) {
	t.Run("fetch", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		src := &fakeSource{fetchErr: errors.New("broker unreachable"), cancel: cancel}

		err := Run(ctx, src, NewHandler(&fakeStore{}, &fakeDLQ{}, Backoff{}))

		require.ErrorContains(t, err, "broker unreachable")
	})
	t.Run("commit", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		src := &fakeSource{queue: []kafka.Message{message(5, validEvent)}, commitErr: errors.New("group rebalancing"), cancel: cancel}

		err := Run(ctx, src, NewHandler(&fakeStore{}, &fakeDLQ{}, Backoff{}))

		require.ErrorContains(t, err, "group rebalancing")
		assert.ErrorContains(t, err, "offset 5")
	})
}

func TestLoadConfig(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

	t.Run("defaults", func(t *testing.T) {
		cfg, err := LoadConfig(env(map[string]string{"MONGO_URI": "mongodb://mongo:27017"}))
		require.NoError(t, err)
		assert.Equal(t, []string{"localhost:9092"}, cfg.Kafka.Brokers)
		assert.Equal(t, "activity-log.events", cfg.Kafka.Topic)
		assert.Equal(t, "activity-log-writer", cfg.Kafka.GroupID)
		assert.Equal(t, "activity-log.events.dlq", cfg.Kafka.DLQTopic)
		assert.Equal(t, "activity_log", cfg.MongoDB)
	})
	t.Run("overrides and a broker list", func(t *testing.T) {
		cfg, err := LoadConfig(env(map[string]string{
			"MONGO_URI": "mongodb://m", "MONGO_DB": "audit", "KAFKA_BROKERS": " a:9092, b:9092 ,", "KAFKA_GROUP": "g",
		}))
		require.NoError(t, err)
		assert.Equal(t, []string{"a:9092", "b:9092"}, cfg.Kafka.Brokers)
		assert.Equal(t, "audit", cfg.MongoDB)
		assert.Equal(t, "g", cfg.Kafka.GroupID)
	})
	t.Run("mongo uri is required", func(t *testing.T) {
		_, err := LoadConfig(env(nil))
		assert.ErrorContains(t, err, "MONGO_URI")
	})
	t.Run("a broker list without brokers", func(t *testing.T) {
		_, err := LoadConfig(env(map[string]string{"MONGO_URI": "mongodb://m", "KAFKA_BROKERS": " , "}))
		assert.Error(t, err)
	})
}
