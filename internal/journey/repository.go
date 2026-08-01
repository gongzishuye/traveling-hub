package journey

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository owns the SQL required to make the daily journey rhythm durable.
// It intentionally has no timer: callers reconcile a frog on reads or in a
// shared worker, and PostgreSQL row locks make those calls idempotent.
type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) Repository { return Repository{db: db} }

// CreateIfAbsent persists one immutable daily plan. On a unique-key race it
// returns the already-locked plan instead of producing a new random result.
func (r Repository) CreateIfAbsent(ctx context.Context, planned DailyJourney) (DailyJourney, bool, error) {
	if err := validatePlannedJourney(planned); err != nil {
		return DailyJourney{}, false, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return DailyJourney{}, false, fmt.Errorf("begin journey creation: %w", err)
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, `
		INSERT INTO daily_journeys
			(id, frog_id, local_date, status, template_id, postcard_id, food_id, departed_at, return_at, next_stage)
		VALUES ($1, $2, $3::date, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (frog_id, local_date) DO NOTHING
		RETURNING id, frog_id, local_date, status, template_id, postcard_id, food_id, departed_at, return_at, next_stage, returned_at`,
		planned.ID, planned.FrogID, planned.LocalDate.In(shanghai).Format("2006-01-02"), planned.Status,
		planned.TemplateID, planned.PostcardID, planned.FoodID, planned.DepartedAt.UTC(), planned.ReturnAt.UTC(), planned.NextStage,
	)
	stored, err := scanDailyJourney(row)
	if err == nil {
		if err := appendLifecycleEvent(ctx, tx, stored, 0); err != nil {
			return DailyJourney{}, false, err
		}
		if err := tx.Commit(); err != nil {
			return DailyJourney{}, false, fmt.Errorf("commit journey creation: %w", err)
		}
		return stored, true, nil
	}
	if err != sql.ErrNoRows {
		return DailyJourney{}, false, fmt.Errorf("insert daily journey: %w", err)
	}
	stored, err = loadForDateTx(ctx, tx, planned.FrogID, planned.LocalDate)
	if err != nil {
		return DailyJourney{}, false, fmt.Errorf("load existing daily journey: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return DailyJourney{}, false, fmt.Errorf("commit existing journey load: %w", err)
	}
	return stored, false, nil
}

// AdvanceDue records all passed narrative checkpoints and the final return.
// A final return and its album entry are deliberately committed together.
func (r Repository) AdvanceDue(ctx context.Context, frogID uuid.UUID, now time.Time) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin journey transition: %w", err)
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, `
		SELECT id, frog_id, local_date, status, template_id, postcard_id, food_id, departed_at, return_at, next_stage, returned_at
		FROM daily_journeys
		WHERE frog_id = $1 AND status = $2 AND return_at <= $3
		ORDER BY return_at ASC
		LIMIT 1
		FOR UPDATE`, frogID, Travelling, now.UTC())
	journey, err := scanDailyJourney(row)
	if err == sql.ErrNoRows {
		if err := tx.Commit(); err != nil {
			return false, fmt.Errorf("commit empty journey transition: %w", err)
		}
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("lock due journey: %w", err)
	}
	for stage := journey.NextStage; stage <= 2; stage++ {
		if err := appendLifecycleEvent(ctx, tx, journey, stage); err != nil {
			return false, err
		}
	}

	returnedAt := now.UTC()
	if _, err := tx.ExecContext(ctx, `
		UPDATE daily_journeys
		SET status = $2, next_stage = 3, returned_at = $3, updated_at = NOW()
		WHERE id = $1`, journey.ID, Returned, returnedAt); err != nil {
		return false, fmt.Errorf("mark journey returned: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO album_entries (id, frog_id, journey_id, postcard_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (journey_id) DO NOTHING`, uuid.New(), journey.FrogID, journey.ID, journey.PostcardID); err != nil {
		return false, fmt.Errorf("insert album entry: %w", err)
	}
	if err := appendLifecycleEvent(ctx, tx, journey, 3); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit journey transition: %w", err)
	}
	return true, nil
}

// AdvanceNarrativeStages keeps next_stage durable for workers and future
// event delivery. The snapshot independently derives visibility from those
// locked timestamps, so an interrupted process cannot hide a passed stage.
func (r Repository) AdvanceNarrativeStages(ctx context.Context, frogID uuid.UUID, now time.Time) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin narrative transition: %w", err)
	}
	defer tx.Rollback()
	row := tx.QueryRowContext(ctx, `
		SELECT id, frog_id, local_date, status, template_id, postcard_id, food_id, departed_at, return_at, next_stage, returned_at
		FROM daily_journeys
		WHERE frog_id = $1 AND status = $2
		ORDER BY local_date DESC
		LIMIT 1
		FOR UPDATE`, frogID, Travelling)
	journey, err := scanDailyJourney(row)
	if err == sql.ErrNoRows {
		if err := tx.Commit(); err != nil {
			return false, fmt.Errorf("commit empty narrative transition: %w", err)
		}
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("lock travelling journey: %w", err)
	}
	if !now.Before(journey.ReturnAt) {
		if err := tx.Rollback(); err != nil {
			return false, fmt.Errorf("rollback final transition: %w", err)
		}
		return r.AdvanceDue(ctx, frogID, now)
	}
	nextStage := journey.NextStage
	for nextStage <= 2 && !now.Before(stageDueAt(journey, nextStage)) {
		if err := appendLifecycleEvent(ctx, tx, journey, nextStage); err != nil {
			return false, err
		}
		nextStage++
	}
	if nextStage == journey.NextStage {
		if err := tx.Commit(); err != nil {
			return false, fmt.Errorf("commit unchanged narrative transition: %w", err)
		}
		return false, nil
	}
	if _, err := tx.ExecContext(ctx, `UPDATE daily_journeys SET next_stage = $2, updated_at = NOW() WHERE id = $1`, journey.ID, nextStage); err != nil {
		return false, fmt.Errorf("advance narrative stage: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit narrative transition: %w", err)
	}
	return true, nil
}

func (r Repository) LatestForDisplay(ctx context.Context, frogID uuid.UUID, localDay string) (*DailyJourney, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, frog_id, local_date, status, template_id, postcard_id, food_id, departed_at, return_at, next_stage, returned_at
		FROM daily_journeys
		WHERE frog_id = $1 AND (local_date = $2::date OR status = $3)
		ORDER BY (local_date = $2::date) DESC, local_date DESC
		LIMIT 1`, frogID, localDay, Returned)
	journey, err := scanDailyJourney(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load display journey: %w", err)
	}
	return &journey, nil
}

func (r Repository) ForDate(ctx context.Context, frogID uuid.UUID, date time.Time) (*DailyJourney, error) {
	journey, err := r.loadForDate(ctx, frogID, date)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load daily journey: %w", err)
	}
	return &journey, nil
}

func (r Repository) AlbumPostcardIDs(ctx context.Context, frogID uuid.UUID) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT postcard_id FROM album_entries
		WHERE frog_id = $1
		ORDER BY created_at, id`, frogID)
	if err != nil {
		return nil, fmt.Errorf("list album entries: %w", err)
	}
	defer rows.Close()
	result := []string{}
	for rows.Next() {
		var postcardID string
		if err := rows.Scan(&postcardID); err != nil {
			return nil, fmt.Errorf("scan album entry: %w", err)
		}
		result = append(result, postcardID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate album entries: %w", err)
	}
	return result, nil
}

// FrogIDsAfter returns a stable, bounded scan of traveler IDs. The caller
// owns the scan cursor; returning an empty slice signals that the current pass
// is complete and the next cycle may start again from the beginning.
func (r Repository) FrogIDsAfter(ctx context.Context, after uuid.UUID, limit int) ([]uuid.UUID, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("worker batch limit must be positive")
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id FROM frogs
		WHERE ($1::uuid IS NULL OR id > $1)
		ORDER BY id
		LIMIT $2`, nullableUUID(after), limit)
	if err != nil {
		return nil, fmt.Errorf("list frogs for journey worker: %w", err)
	}
	defer rows.Close()
	ids := make([]uuid.UUID, 0, limit)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan frog for journey worker: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate frogs for journey worker: %w", err)
	}
	return ids, nil
}

func (r Repository) FrogExistedBefore(ctx context.Context, frogID uuid.UUID, cutoff time.Time) (bool, error) {
	var exists bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM frogs WHERE id = $1 AND created_at < $2)`, frogID, cutoff.UTC()).Scan(&exists); err != nil {
		return false, fmt.Errorf("check frog creation time: %w", err)
	}
	return exists, nil
}

func nullableUUID(value uuid.UUID) any {
	if value == uuid.Nil {
		return nil
	}
	return value
}

func (r Repository) loadForDate(ctx context.Context, frogID uuid.UUID, date time.Time) (DailyJourney, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, frog_id, local_date, status, template_id, postcard_id, food_id, departed_at, return_at, next_stage, returned_at
		FROM daily_journeys WHERE frog_id = $1 AND local_date = $2::date`, frogID, date.In(shanghai).Format("2006-01-02"))
	return scanDailyJourney(row)
}

func loadForDateTx(ctx context.Context, tx *sql.Tx, frogID uuid.UUID, date time.Time) (DailyJourney, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT id, frog_id, local_date, status, template_id, postcard_id, food_id, departed_at, return_at, next_stage, returned_at
		FROM daily_journeys WHERE frog_id = $1 AND local_date = $2::date`, frogID, date.In(shanghai).Format("2006-01-02"))
	return scanDailyJourney(row)
}

// appendLifecycleEvent writes an Agent-visible event in the same transaction
// as the state it describes. The deduplication key is a second safety net for
// retries; the journey row lock normally prevents that path from being used.
func appendLifecycleEvent(ctx context.Context, tx *sql.Tx, journey DailyJourney, stage int) error {
	if stage < 0 || stage > 3 {
		return fmt.Errorf("invalid journey lifecycle stage %d", stage)
	}
	key := fmt.Sprintf("journey:%s:stage:%d", journey.ID, stage)
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM events WHERE deduplication_key = $1)`, key).Scan(&exists); err != nil {
		return fmt.Errorf("check journey event deduplication: %w", err)
	}
	if exists {
		return nil
	}

	var agentID uuid.UUID
	if err := tx.QueryRowContext(ctx, `SELECT agent_id FROM frogs WHERE id = $1`, journey.FrogID).Scan(&agentID); err != nil {
		return fmt.Errorf("load journey event audience: %w", err)
	}
	template, found := TemplateByID(journey.TemplateID)
	if !found {
		return fmt.Errorf("load journey event template: %s", journey.TemplateID)
	}
	var typeName, text string
	var occurredAt time.Time
	switch stage {
	case 0:
		typeName, text, occurredAt = "journey.departed", departureText, journey.DepartedAt
	case 1, 2:
		typeName, text, occurredAt = "journey.stage", template.Events[stage-1], stageDueAt(journey, stage)
	case 3:
		typeName, text, occurredAt = "journey.returned", template.Events[2], journey.ReturnAt
	}
	payload, err := json.Marshal(map[string]any{
		"journey_id": journey.ID.String(),
		"frog_id":    journey.FrogID.String(),
		"stage":      stage,
		"text":       text,
	})
	if err != nil {
		return fmt.Errorf("encode journey event payload: %w", err)
	}
	var worldVersion int64
	if err := tx.QueryRowContext(ctx, `
		UPDATE world_state SET version = version + 1, updated_at = $1
		WHERE singleton = TRUE
		RETURNING version`, occurredAt.UTC()).Scan(&worldVersion); err != nil {
		return fmt.Errorf("advance world for journey event: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO world_ticks (version, advanced_at) VALUES ($1, $2)`, worldVersion, occurredAt.UTC()); err != nil {
		return fmt.Errorf("record journey event world tick: %w", err)
	}
	eventID := uuid.New()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO events
			(id, type, world_version, occurred_at, payload, source, journey_id, journey_stage, deduplication_key)
		VALUES ($1, $2, $3, $4, $5, 'journey', $6, $7, $8)`,
		eventID, typeName, worldVersion, occurredAt.UTC(), payload, journey.ID, stage, key); err != nil {
		return fmt.Errorf("insert journey lifecycle event: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO event_audience (event_id, agent_id) VALUES ($1, $2)`, eventID, agentID); err != nil {
		return fmt.Errorf("insert journey event audience: %w", err)
	}
	return nil
}

type rowScanner interface{ Scan(...any) error }

func scanDailyJourney(row rowScanner) (DailyJourney, error) {
	var journey DailyJourney
	var returnedAt sql.NullTime
	err := row.Scan(
		&journey.ID, &journey.FrogID, &journey.LocalDate, &journey.Status,
		&journey.TemplateID, &journey.PostcardID, &journey.FoodID,
		&journey.DepartedAt, &journey.ReturnAt, &journey.NextStage, &returnedAt,
	)
	if err != nil {
		return DailyJourney{}, err
	}
	journey.DepartedAt = journey.DepartedAt.UTC()
	journey.ReturnAt = journey.ReturnAt.UTC()
	if returnedAt.Valid {
		value := returnedAt.Time.UTC()
		journey.ReturnedAt = &value
	}
	return journey, nil
}

func validatePlannedJourney(journey DailyJourney) error {
	if journey.ID == uuid.Nil || journey.FrogID == uuid.Nil || journey.Status != Travelling || journey.NextStage != 1 || journey.DepartedAt.IsZero() || !journey.ReturnAt.After(journey.DepartedAt) {
		return fmt.Errorf("invalid planned journey")
	}
	template, found := TemplateByID(journey.TemplateID)
	if !found || template.PostcardID != journey.PostcardID || template.FoodID != journey.FoodID {
		return fmt.Errorf("invalid journey template")
	}
	return nil
}
