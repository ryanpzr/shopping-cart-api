package outbox

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ryanpzr/shopping-cart-api/pkg/database"
)

const queryPurge = `DELETE FROM outbox_events WHERE created_at < $1`

// Purge deletes the events created before the cutoff and returns how many.
// Debezium reads the inserts from the WAL, so old rows are no longer needed;
// the retention window only keeps a few days for debugging and replays.
func Purge(ctx context.Context, db database.DBTX, before time.Time) (int64, error) {
	result, err := db.ExecContext(ctx, queryPurge, before)
	if err != nil {
		return 0, fmt.Errorf("outbox: purge: %w", err)
	}
	return result.RowsAffected()
}

// StartPurger deletes the events older than retention every interval until
// ctx is cancelled. It runs in its own goroutine and logs failures instead of
// stopping, because a missed purge only delays the cleanup. The DELETE is
// idempotent, so several API instances can run it at the same time.
func StartPurger(ctx context.Context, db database.DBTX, interval, retention time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				deleted, err := Purge(ctx, db, time.Now().Add(-retention))
				switch {
				case ctx.Err() != nil:
					return
				case err != nil:
					log.Printf("outbox purge failed: %v", err)
				case deleted > 0:
					log.Printf("outbox purge: deleted %d events older than %s", deleted, retention)
				}
			}
		}
	}()
}
