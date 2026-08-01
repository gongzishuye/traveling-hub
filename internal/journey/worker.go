package journey

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

const workerAdvisoryLockID int64 = 884_461_952_107

// Worker advances the shared world from one bounded scan. It never owns a
// timer per traveler. PostgreSQL's advisory lock means only one app instance
// performs a cycle at a time; service reconciliation remains idempotent as a
// second line of defence.
type Worker struct {
	db       *sql.DB
	service  Service
	interval time.Duration
	batch    int

	mu     sync.Mutex
	cursor uuid.UUID
}

func NewWorker(db *sql.DB, service Service, interval time.Duration, batch int) (*Worker, error) {
	if db == nil {
		return nil, fmt.Errorf("journey worker database is required")
	}
	if interval <= 0 {
		return nil, fmt.Errorf("journey worker interval must be positive")
	}
	if batch <= 0 {
		return nil, fmt.Errorf("journey worker batch size must be positive")
	}
	return &Worker{db: db, service: service, interval: interval, batch: batch}, nil
}

// Run performs an immediate catch-up cycle, then continues until its context
// is cancelled. A failed cycle is intentionally retried on the next interval.
func (w *Worker) Run(ctx context.Context) {
	_ = w.Cycle(ctx)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = w.Cycle(ctx)
		}
	}
}

// Cycle holds a database-wide advisory lock while it reconciles at most one
// batch. An unavailable lock is normal when another replica is the worker.
func (w *Worker) Cycle(ctx context.Context) error {
	conn, err := w.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("open journey worker connection: %w", err)
	}
	defer conn.Close()

	var locked bool
	if err := conn.QueryRowContext(ctx, `SELECT pg_try_advisory_lock($1)`, workerAdvisoryLockID).Scan(&locked); err != nil {
		return fmt.Errorf("acquire journey worker lock: %w", err)
	}
	if !locked {
		return nil
	}
	defer func() {
		_, _ = conn.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, workerAdvisoryLockID)
	}()

	w.mu.Lock()
	defer w.mu.Unlock()
	ids, err := w.service.repository.FrogIDsAfter(ctx, w.cursor, w.batch)
	if err != nil {
		return err
	}
	if len(ids) == 0 && w.cursor != uuid.Nil {
		w.cursor = uuid.Nil
		ids, err = w.service.repository.FrogIDsAfter(ctx, w.cursor, w.batch)
		if err != nil {
			return err
		}
	}
	if len(ids) == 0 {
		return nil
	}
	now := w.service.clock.Now()
	for _, frogID := range ids {
		if _, err := w.service.ReconcileAt(ctx, frogID, now); err != nil {
			return fmt.Errorf("reconcile journey for frog %s: %w", frogID, err)
		}
	}
	w.cursor = ids[len(ids)-1]
	return nil
}
