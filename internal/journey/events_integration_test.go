//go:build integration

package journey

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/faria/traveling-hub/internal/event"
	postgresplatform "github.com/faria/traveling-hub/internal/platform/postgres"
	"github.com/google/uuid"
)

func TestReconcilePublishesOneAgentVisibleEventForEachJourneyLifecycleStep(t *testing.T) {
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
	var agentID uuid.UUID
	if err := db.QueryRow(`SELECT agent_id FROM frogs WHERE id = $1`, frogID).Scan(&agentID); err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(db), fixedClock{}, SelectorFunc(func(uuid.UUID, time.Time) (int, int, error) {
		return 0, 120, nil // 20:00 Shanghai
	}))

	for _, now := range []time.Time{
		time.Date(2026, 8, 1, 8, 0, 0, 0, shanghai),
		time.Date(2026, 8, 1, 11, 0, 0, 0, shanghai),
		time.Date(2026, 8, 1, 14, 0, 0, 0, shanghai),
		time.Date(2026, 8, 1, 20, 1, 0, 0, shanghai),
	} {
		if _, err := service.ReconcileAt(ctx, frogID, now); err != nil {
			t.Fatal(err)
		}
	}
	// A retry after return must not append a second return event or card.
	if _, err := service.ReconcileAt(ctx, frogID, time.Date(2026, 8, 1, 21, 0, 0, 0, shanghai)); err != nil {
		t.Fatal(err)
	}

	events, _, err := event.NewService(db).ListAfter(ctx, agentID, "", 20)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(events), 4; got != want {
		t.Fatalf("visible lifecycle event count = %d, want %d: %#v", got, want, events)
	}
	for index, want := range []string{"journey.departed", "journey.stage", "journey.stage", "journey.returned"} {
		if events[index].Type != want {
			t.Errorf("event[%d].Type = %q, want %q", index, events[index].Type, want)
		}
		if events[index].WorldVersion < 1 {
			t.Errorf("event[%d].WorldVersion = %d, want positive", index, events[index].WorldVersion)
		}
	}
	var cards int
	if err := db.QueryRow(`SELECT count(*) FROM album_entries WHERE frog_id = $1`, frogID).Scan(&cards); err != nil {
		t.Fatal(err)
	}
	if cards != 1 {
		t.Fatalf("album entry count = %d, want 1", cards)
	}
}

func TestLateReconcileBackfillsEveryMissingJourneyLifecycleEvent(t *testing.T) {
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
	var agentID uuid.UUID
	if err := db.QueryRow(`SELECT agent_id FROM frogs WHERE id = $1`, frogID).Scan(&agentID); err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(db), fixedClock{}, SelectorFunc(func(uuid.UUID, time.Time) (int, int, error) {
		return 0, 120, nil
	}))
	if _, err := service.ReconcileAt(ctx, frogID, time.Date(2026, 8, 1, 8, 0, 0, 0, shanghai)); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ReconcileAt(ctx, frogID, time.Date(2026, 8, 1, 20, 1, 0, 0, shanghai)); err != nil {
		t.Fatal(err)
	}

	events, _, err := event.NewService(db).ListAfter(ctx, agentID, "", 20)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(events), 4; got != want {
		t.Fatalf("late reconciliation lifecycle event count = %d, want %d: %#v", got, want, events)
	}
}
