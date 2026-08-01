package frog

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type Frog struct {
	ID      uuid.UUID `json:"id"`
	AgentID uuid.UUID `json:"agent_id"`
}

type Service struct{ db *sql.DB }

func NewService(db *sql.DB) Service { return Service{db: db} }

func (s Service) GetForAgent(ctx context.Context, agentID uuid.UUID) (Frog, error) {
	var frog Frog
	err := s.db.QueryRowContext(ctx, `SELECT id, agent_id FROM frogs WHERE agent_id = $1`, agentID).Scan(&frog.ID, &frog.AgentID)
	if err != nil {
		return Frog{}, fmt.Errorf("load frog: %w", err)
	}
	return frog, nil
}

// GetForUser resolves the one traveler that belongs to the authenticated Web
// user. The relationship is derived server-side so browser requests never get
// to choose an agent or frog identifier.
func (s Service) GetForUser(ctx context.Context, userID uuid.UUID) (Frog, error) {
	var frog Frog
	err := s.db.QueryRowContext(ctx, `
		SELECT f.id, f.agent_id
		FROM frogs f
		JOIN agents a ON a.id = f.agent_id
		WHERE a.user_id = $1`, userID).Scan(&frog.ID, &frog.AgentID)
	if err != nil {
		return Frog{}, fmt.Errorf("load frog for user: %w", err)
	}
	return frog, nil
}

func (s Service) EnsureForAgent(ctx context.Context, agentID uuid.UUID) (Frog, error) {
	var frog Frog
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO frogs (id, agent_id) VALUES ($1, $2)
		ON CONFLICT (agent_id) DO UPDATE SET agent_id = EXCLUDED.agent_id
		RETURNING id, agent_id`, uuid.New(), agentID).Scan(&frog.ID, &frog.AgentID)
	if err != nil {
		return Frog{}, fmt.Errorf("create or load frog: %w", err)
	}
	return frog, nil
}

func InsertForAgentTx(ctx context.Context, tx *sql.Tx, agentID uuid.UUID) (Frog, error) {
	frog := Frog{ID: uuid.New(), AgentID: agentID}
	if _, err := tx.ExecContext(ctx, `INSERT INTO frogs (id, agent_id) VALUES ($1, $2)`, frog.ID, frog.AgentID); err != nil {
		return Frog{}, fmt.Errorf("insert frog: %w", err)
	}
	return frog, nil
}
