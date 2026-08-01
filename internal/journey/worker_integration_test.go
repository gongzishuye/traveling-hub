//go:build integration

package journey

import (
	"context"
	"os"
	"testing"
	"time"

	postgresplatform "github.com/faria/traveling-hub/internal/platform/postgres"
	"github.com/google/uuid"
)

type workerClock struct{ now time.Time }

func (c *workerClock) Now() time.Time { return c.now }

func TestWorkerCycleCreatesAndReturnsOfflineJourneys(t *testing.T) {
	dsn := os.Getenv("TRAVELINGHUB_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TRAVELINGHUB_POSTGRES_DSN is required for integration tests")
	}
	ctx := context.Background()
	db, err := postgresplatform.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := postgresplatform.ApplyMigrations(ctx, db, "migrations"); err != nil {
		t.Fatal(err)
	}
	resetJourneyTables(t, db)
	frogID := insertTestFrog(t, db)
	clock := &workerClock{now: time.Date(2026, 8, 1, 8, 0, 0, 0, shanghai)}
	service := NewService(NewRepository(db), clock, SelectorFunc(func(uuid.UUID, time.Time) (int, int, error) {
		return 0, 120, nil
	}))
	worker, err := NewWorker(db, service, time.Minute, 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := worker.Cycle(ctx); err != nil {
		t.Fatal(err)
	}
	assertCount(t, db, `SELECT count(*) FROM daily_journeys WHERE frog_id = $1`, frogID, 1)

	var returnAt time.Time
	if err := db.QueryRow(`SELECT return_at FROM daily_journeys WHERE frog_id = $1`, frogID).Scan(&returnAt); err != nil {
		t.Fatal(err)
	}
	clock.now = returnAt.Add(time.Second)
	if err := worker.Cycle(ctx); err != nil {
		t.Fatal(err)
	}
	assertCount(t, db, `SELECT count(*) FROM album_entries WHERE frog_id = $1`, frogID, 1)
	if err := worker.Cycle(ctx); err != nil {
		t.Fatal(err)
	}
	assertCount(t, db, `SELECT count(*) FROM album_entries WHERE frog_id = $1`, frogID, 1)
}
