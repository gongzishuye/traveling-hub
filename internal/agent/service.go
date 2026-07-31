package agent

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Principal struct {
	AgentID uuid.UUID
}

type Service struct{ db *sql.DB }

func NewService(db *sql.DB) Service { return Service{db: db} }

func (s Service) Authenticate(ctx context.Context, rawToken string) (Principal, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return Principal{}, fmt.Errorf("missing bearer token")
	}
	digest := sha256.Sum256([]byte(rawToken))
	id := uuid.New()
	var agentID uuid.UUID
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO agents (id, token_digest) VALUES ($1, $2)
		ON CONFLICT (token_digest) DO UPDATE SET updated_at = NOW()
		RETURNING id`, id, digest[:]).Scan(&agentID)
	if err != nil {
		return Principal{}, fmt.Errorf("create or load agent: %w", err)
	}
	return Principal{AgentID: agentID}, nil
}
