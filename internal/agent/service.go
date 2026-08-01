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
	var agentID uuid.UUID
	column := "token_digest"
	if strings.HasPrefix(rawToken, "thub_") {
		column = "api_key_digest"
	}
	err := s.db.QueryRowContext(ctx, `SELECT id FROM agents WHERE `+column+` = $1`, digest[:]).Scan(&agentID)
	if err != nil {
		return Principal{}, fmt.Errorf("authenticate agent: %w", err)
	}
	return Principal{AgentID: agentID}, nil
}
