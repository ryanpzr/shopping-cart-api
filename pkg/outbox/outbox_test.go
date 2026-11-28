package outbox_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/pkg/outbox"
)

// fakeDB records the statements it receives. It satisfies database.DBTX.
type fakeDB struct {
	execs    []call
	execErr  error
	affected int64
}

type call struct {
	query string
	args  []any
}

type fakeResult struct{ affected int64 }

func (r fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (r fakeResult) RowsAffected() (int64, error) { return r.affected, nil }

func (f *fakeDB) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	f.execs = append(f.execs, call{query, args})
	return fakeResult{f.affected}, f.execErr
}

func (f *fakeDB) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return nil, errors.New("not used")
}

func (f *fakeDB) QueryRowContext(context.Context, string, ...any) *sql.Row { return nil }

func TestInsert_WritesTheOutboxRow(t *testing.T) {
	db := &fakeDB{}
	occurred := time.Date(2026, 11, 28, 10, 30, 0, 0, time.UTC)
	event := outbox.Event{
		ID:         uuid.MustParse("0b6f1c2e-1111-4222-8333-444455556666"),
		UserID:     42,
		Type:       outbox.EventCartCheckout,
		Metadata:   map[string]any{"order_id": 10, "total_price": 179.8},
		OccurredAt: occurred,
	}

	require.NoError(t, outbox.NewWriter(db).Insert(context.Background(), event))

	require.Len(t, db.execs, 1)
	args := db.execs[0].args
	assert.Equal(t, event.ID, args[0])
	assert.Equal(t, "cart", args[1], "aggregate type")
	assert.Equal(t, "42", args[2], "aggregate id is the user id: it is the Kafka key")
	assert.Equal(t, "cart_checkout", args[3])

	var msg outbox.Message
	require.NoError(t, json.Unmarshal([]byte(args[4].(string)), &msg))
	assert.Equal(t, "0b6f1c2e-1111-4222-8333-444455556666", msg.EventID)
	assert.Equal(t, 42, msg.UserID)
	assert.Equal(t, "cart_checkout", msg.EventType)
	assert.Equal(t, occurred, msg.OccurredAt)
	assert.EqualValues(t, 10, msg.Metadata["order_id"])
	assert.Equal(t, 179.8, msg.Metadata["total_price"])
}

func TestInsert_NilMetadataBecomesAnEmptyObject(t *testing.T) {
	db := &fakeDB{}

	require.NoError(t, outbox.NewWriter(db).Insert(context.Background(), outbox.New(7, outbox.EventLogin, nil)))

	assert.Contains(t, db.execs[0].args[4], `"metadata":{}`)
}

func TestInsert_AggregateTypes(t *testing.T) {
	want := map[string]string{
		outbox.EventLogin:          "user",
		outbox.EventProductCreated: "product",
		outbox.EventProductUpdated: "product",
		outbox.EventProductDeleted: "product",
		outbox.EventCartCheckout:   "cart",
		outbox.EventOrderPlaced:    "order",
		outbox.EventOrderCancelled: "order",
	}
	require.Len(t, outbox.EventTypes, len(want))
	for eventType, aggregate := range want {
		assert.True(t, outbox.ValidType(eventType), eventType)
		assert.Equal(t, aggregate, outbox.AggregateType(eventType), eventType)
	}
	assert.False(t, outbox.ValidType("user_banned"))
	assert.False(t, outbox.ValidType(""))
}

func TestInsert_RejectsInvalidEventsWithoutTouchingTheDatabase(t *testing.T) {
	valid := outbox.New(7, outbox.EventLogin, nil)
	tests := map[string]func(e *outbox.Event){
		"nil id":        func(e *outbox.Event) { e.ID = uuid.Nil },
		"no user":       func(e *outbox.Event) { e.UserID = 0 },
		"negative user": func(e *outbox.Event) { e.UserID = -1 },
		"unknown type":  func(e *outbox.Event) { e.Type = "user_banned" },
		"no time":       func(e *outbox.Event) { e.OccurredAt = time.Time{} },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			db := &fakeDB{}
			e := valid
			mutate(&e)

			err := outbox.NewWriter(db).Insert(context.Background(), e)

			require.Error(t, err)
			assert.Empty(t, db.execs)
		})
	}
}

func TestInsert_DatabaseFailureIsReturned(t *testing.T) {
	// The caller rolls its transaction back when this fails: an action
	// without its audit event is not allowed.
	db := &fakeDB{execErr: errors.New("connection reset")}

	err := outbox.NewWriter(db).Insert(context.Background(), outbox.New(7, outbox.EventLogin, nil))

	require.Error(t, err)
	assert.ErrorContains(t, err, "connection reset")
}

func TestNew_GeneratesIDAndUTCTime(t *testing.T) {
	a := outbox.New(1, outbox.EventLogin, nil)
	b := outbox.New(1, outbox.EventLogin, nil)

	assert.NotEqual(t, a.ID, b.ID)
	assert.Equal(t, time.UTC, a.OccurredAt.Location())
	assert.WithinDuration(t, time.Now(), a.OccurredAt, 2*time.Second)
	require.NoError(t, a.Validate())
}

func TestPurge_DeletesOlderThanTheCutoff(t *testing.T) {
	db := &fakeDB{affected: 5}
	cutoff := time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC)

	deleted, err := outbox.Purge(context.Background(), db, cutoff)

	require.NoError(t, err)
	assert.EqualValues(t, 5, deleted)
	require.Len(t, db.execs, 1)
	assert.Contains(t, db.execs[0].query, "DELETE FROM outbox_events")
	assert.Equal(t, cutoff, db.execs[0].args[0])
}

func TestPurge_Failure(t *testing.T) {
	_, err := outbox.Purge(context.Background(), &fakeDB{execErr: errors.New("boom")}, time.Now())
	assert.Error(t, err)
}

func TestStartPurger_RunsUntilCancelled(t *testing.T) {
	db := &syncDB{ran: make(chan struct{}, 4)}
	ctx, cancel := context.WithCancel(context.Background())

	outbox.StartPurger(ctx, db, 5*time.Millisecond, time.Hour)

	select {
	case <-db.ran:
	case <-time.After(2 * time.Second):
		t.Fatal("the purger did not run")
	}
	cancel()
	time.Sleep(30 * time.Millisecond) // let a pending tick finish
	runs := len(db.ran)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, runs, len(db.ran), "no purge after cancellation")
}

// syncDB signals every purge on a channel; safe to call from the purger goroutine.
type syncDB struct{ ran chan struct{} }

func (s *syncDB) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	select {
	case s.ran <- struct{}{}:
	default:
	}
	return fakeResult{}, nil
}

func (s *syncDB) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return nil, errors.New("not used")
}

func (s *syncDB) QueryRowContext(context.Context, string, ...any) *sql.Row { return nil }
