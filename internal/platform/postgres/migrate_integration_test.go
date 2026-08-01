//go:build integration

package postgres

import (
	"context"
	"os"
	"testing"
)

func TestApplyMigrationsReleasesItsAdvisoryLock(t *testing.T) {
	dsn := os.Getenv("TRAVELINGHUB_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TRAVELINGHUB_POSTGRES_DSN is required for integration tests")
	}
	ctx := context.Background()
	db, err := Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := ApplyMigrations(ctx, db, "migrations"); err != nil {
		t.Fatal(err)
	}
	other, err := Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = other.Close() })
	var acquired bool
	if err := other.QueryRowContext(ctx, `SELECT pg_try_advisory_lock(742901)`).Scan(&acquired); err != nil {
		t.Fatal(err)
	}
	if !acquired {
		t.Fatal("migration advisory lock remained held after ApplyMigrations")
	}
	if _, err := other.ExecContext(ctx, `SELECT pg_advisory_unlock(742901)`); err != nil {
		t.Fatal(err)
	}
}
