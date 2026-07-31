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
