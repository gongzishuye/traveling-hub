package world

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type State struct {
	Version    int64
	AdvancedAt time.Time
}

type Service struct {
	db    *sql.DB
	clock Clock
}

func NewService(db *sql.DB, clock Clock) Service { return Service{db: db, clock: clock} }

func (s Service) Advance(ctx context.Context) (State, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return State{}, fmt.Errorf("begin world tick: %w", err)
	}
	defer tx.Rollback()
	now := s.clock.Now().UTC()
	var version int64
	if err := tx.QueryRowContext(ctx, `UPDATE world_state SET version = version + 1, updated_at = $1 WHERE singleton = TRUE RETURNING version`, now).Scan(&version); err != nil {
		return State{}, fmt.Errorf("advance world version: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO world_ticks (version, advanced_at) VALUES ($1, $2)`, version, now); err != nil {
		return State{}, fmt.Errorf("record world tick: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return State{}, fmt.Errorf("commit world tick: %w", err)
	}
	return State{Version: version, AdvancedAt: now}, nil
}
