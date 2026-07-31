package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func ApplyMigrations(ctx context.Context, db *sql.DB, directory string) error {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}
	directory, err := resolveMigrationDirectory(directory)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("read migration directory: %w", err)
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		var exists bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, name).Scan(&exists); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if exists {
			continue
		}
		body, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}
		if _, err = tx.ExecContext(ctx, string(body)); err == nil {
			_, err = tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name)
		}
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}
	return nil
}

func resolveMigrationDirectory(directory string) (string, error) {
	if filepath.IsAbs(directory) {
		if _, err := os.Stat(directory); err != nil {
			return "", fmt.Errorf("locate migration directory: %w", err)
		}
		return directory, nil
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	for current := workingDirectory; ; current = filepath.Dir(current) {
		candidate := filepath.Join(current, directory)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
	}
	return "", fmt.Errorf("locate migration directory %q from %q", directory, workingDirectory)
}
