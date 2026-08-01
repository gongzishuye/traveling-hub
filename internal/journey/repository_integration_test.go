//go:build integration

package journey

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	postgresplatform "github.com/faria/traveling-hub/internal/platform/postgres"
	"github.com/google/uuid"
)

func TestReconcilePersistsOnePlanAndOneAlbumEntry(t *testing.T) {
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
	service := NewService(NewRepository(db), fixedClock{}, SelectorFunc(func(uuid.UUID, time.Time) (int, int, error) {
		return 0, 120, nil // 20:00 Shanghai
	}))

	beforeDeparture := time.Date(2026, 8, 1, 7, 59, 0, 0, shanghai)
	snapshot, err := service.ReconcileAt(ctx, frogID, beforeDeparture)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Phase != Home || snapshot.Journey != nil {
		t.Fatalf("before departure snapshot = %#v", snapshot)
	}

	departedAt := time.Date(2026, 8, 1, 8, 0, 0, 0, shanghai)
	first, err := service.ReconcileAt(ctx, frogID, departedAt)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.ReconcileAt(ctx, frogID, departedAt.Add(30*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if first.Phase != Travelling || first.Journey == nil || first.Journey.ReturnAt != nil || first.Journey.TemplateID != second.Journey.TemplateID {
		t.Fatalf("travelling snapshots = %#v %#v", first, second)
	}
	assertCount(t, db, `SELECT count(*) FROM daily_journeys WHERE frog_id = $1`, frogID, 1)
	firstCheckpoint, err := service.ReconcileAt(ctx, frogID, time.Date(2026, 8, 1, 11, 0, 0, 0, shanghai))
	if err != nil {
		t.Fatal(err)
	}
	if len(firstCheckpoint.Events) != 2 {
		t.Fatalf("first checkpoint events = %#v", firstCheckpoint.Events)
	}
	assertCount(t, db, `SELECT next_stage FROM daily_journeys WHERE frog_id = $1`, frogID, 2)
	secondCheckpoint, err := service.ReconcileAt(ctx, frogID, time.Date(2026, 8, 1, 14, 0, 0, 0, shanghai))
	if err != nil {
		t.Fatal(err)
	}
	if len(secondCheckpoint.Events) != 3 {
		t.Fatalf("second checkpoint events = %#v", secondCheckpoint.Events)
	}
	assertCount(t, db, `SELECT next_stage FROM daily_journeys WHERE frog_id = $1`, frogID, 3)

	returned, err := service.ReconcileAt(ctx, frogID, time.Date(2026, 8, 1, 20, 1, 0, 0, shanghai))
	if err != nil {
		t.Fatal(err)
	}
	if returned.Phase != Returned || returned.Journey == nil || returned.Journey.ReturnAt == nil || len(returned.Events) != 4 || len(returned.AlbumPostcardIDs) != 1 {
		t.Fatalf("returned snapshot = %#v", returned)
	}
	if _, err := service.ReconcileAt(ctx, frogID, time.Date(2026, 8, 1, 21, 0, 0, 0, shanghai)); err != nil {
		t.Fatal(err)
	}
	assertCount(t, db, `SELECT count(*) FROM album_entries WHERE frog_id = $1`, frogID, 1)
}

func TestReconcileBeforeDepartureBackfillsYesterdayWhenTheWorkerWasOffline(t *testing.T) {
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
	if _, err := db.Exec(`UPDATE frogs SET created_at = $2 WHERE id = $1`, frogID, time.Date(2026, 8, 1, 7, 0, 0, 0, shanghai)); err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(db), fixedClock{}, SelectorFunc(func(uuid.UUID, time.Time) (int, int, error) {
		return 0, 120, nil
	}))
	snapshot, err := service.ReconcileAt(ctx, frogID, time.Date(2026, 8, 2, 7, 0, 0, 0, shanghai))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Phase != Returned || snapshot.Journey == nil || len(snapshot.AlbumPostcardIDs) != 1 {
		t.Fatalf("offline catch-up snapshot = %#v", snapshot)
	}
	assertCount(t, db, `SELECT count(*) FROM daily_journeys WHERE frog_id = $1`, frogID, 1)
}

type fixedClock struct{}

func (fixedClock) Now() time.Time { return time.Time{} }

func resetJourneyTables(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec(`TRUNCATE agent_event_cursors, event_audience, events, album_entries, daily_journeys, world_ticks, frogs, agents, users RESTART IDENTITY CASCADE; UPDATE world_state SET version = 0, updated_at = NOW() WHERE singleton = TRUE`); err != nil {
		t.Fatal(err)
	}
}

func insertTestFrog(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()
	userID, agentID, frogID := uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO users (id, email_normalized, password_hash) VALUES ($1, $2, $3)`, userID, "journey@example.com", "test"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO agents (id, token_digest, api_key_digest, user_id) VALUES ($1, $2, $3, $4)`, agentID, []byte("token"), []byte("api-key"), userID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO frogs (id, agent_id) VALUES ($1, $2)`, frogID, agentID); err != nil {
		t.Fatal(err)
	}
	return frogID
}

func assertCount(t *testing.T, db *sql.DB, query string, argument uuid.UUID, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(query, argument).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("count = %d, want %d", got, want)
	}
}
