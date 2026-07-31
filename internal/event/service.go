package event

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct{ db *sql.DB }

func NewService(db *sql.DB) Service { return Service{db: db} }

func (s Service) Append(ctx context.Context, draft Draft) (Event, error) {
	if !typePattern.MatchString(draft.Type) || draft.WorldVersion < 1 || draft.Source != "fixture" || len(draft.Audience) == 0 || !json.Valid(draft.Payload) {
		return Event{}, fmt.Errorf("invalid event draft")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Event{}, fmt.Errorf("begin append event: %w", err)
	}
	defer tx.Rollback()
	e := Event{ID: uuid.New(), Type: draft.Type, WorldVersion: draft.WorldVersion, OccurredAt: draft.OccurredAt.UTC(), Payload: draft.Payload}
	if _, err := tx.ExecContext(ctx, `INSERT INTO events (id, type, world_version, occurred_at, payload, source) VALUES ($1, $2, $3, $4, $5, $6)`, e.ID, e.Type, e.WorldVersion, e.OccurredAt, e.Payload, draft.Source); err != nil {
		return Event{}, fmt.Errorf("insert event: %w", err)
	}
	for _, agentID := range draft.Audience {
		if _, err := tx.ExecContext(ctx, `INSERT INTO event_audience (event_id, agent_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, e.ID, agentID); err != nil {
			return Event{}, fmt.Errorf("insert event audience: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return Event{}, fmt.Errorf("commit event: %w", err)
	}
	return e, nil
}

func (s Service) ListAfter(ctx context.Context, agentID uuid.UUID, value string, limit int) ([]Event, string, error) {
	if limit < 1 || limit > 100 {
		return nil, "", fmt.Errorf("limit must be between 1 and 100")
	}
	version, at, id, err := s.startCursor(ctx, agentID, value)
	if err != nil {
		return nil, "", err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT e.id, e.type, e.world_version, e.occurred_at, e.payload
		FROM events e JOIN event_audience ea ON ea.event_id = e.id
		WHERE ea.agent_id = $1 AND (
			e.world_version > $2 OR
			(e.world_version = $2 AND e.occurred_at > $3) OR
			(e.world_version = $2 AND e.occurred_at = $3 AND e.id > $4)
		)
		ORDER BY e.world_version, e.occurred_at, e.id
		LIMIT $5`, agentID, version, at, id, limit)
	if err != nil {
		return nil, "", fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()
	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Type, &e.WorldVersion, &e.OccurredAt, &e.Payload); err != nil {
			return nil, "", fmt.Errorf("scan event: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterate events: %w", err)
	}
	if len(events) == 0 {
		return events, "", nil
	}
	return events, encodeCursor(events[len(events)-1]), nil
}

func (s Service) startCursor(ctx context.Context, agentID uuid.UUID, value string) (int64, time.Time, uuid.UUID, error) {
	if value != "" {
		return decodeCursor(value)
	}
	var version int64
	var at time.Time
	var id uuid.UUID
	err := s.db.QueryRowContext(ctx, `SELECT world_version, occurred_at, event_id FROM agent_event_cursors WHERE agent_id = $1`, agentID).Scan(&version, &at, &id)
	if err == sql.ErrNoRows {
		return 0, time.Unix(0, 0).UTC(), uuid.Nil, nil
	}
	if err != nil {
		return 0, time.Time{}, uuid.Nil, fmt.Errorf("load agent cursor: %w", err)
	}
	return version, at, id, nil
}

func (s Service) Acknowledge(ctx context.Context, agentID uuid.UUID, value string) error {
	version, at, id, err := decodeCursor(value)
	if err != nil || id == uuid.Nil {
		return fmt.Errorf("invalid cursor")
	}
	var visible bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM event_audience WHERE event_id = $1 AND agent_id = $2)`, id, agentID).Scan(&visible); err != nil {
		return fmt.Errorf("validate cursor visibility: %w", err)
	}
	if !visible {
		return fmt.Errorf("cursor is not visible to agent")
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO agent_event_cursors (agent_id, event_id, world_version, occurred_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (agent_id) DO UPDATE SET
			event_id = EXCLUDED.event_id,
			world_version = EXCLUDED.world_version,
			occurred_at = EXCLUDED.occurred_at,
			updated_at = NOW()
		WHERE (agent_event_cursors.world_version, agent_event_cursors.occurred_at, agent_event_cursors.event_id)
			< (EXCLUDED.world_version, EXCLUDED.occurred_at, EXCLUDED.event_id)`, agentID, id, version, at)
	if err != nil {
		return fmt.Errorf("acknowledge events: %w", err)
	}
	return nil
}

func ParseLimit(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return 20, nil
	}
	var limit int
	if _, err := fmt.Sscanf(raw, "%d", &limit); err != nil || limit < 1 || limit > 100 {
		return 0, fmt.Errorf("limit must be between 1 and 100")
	}
	return limit, nil
}
