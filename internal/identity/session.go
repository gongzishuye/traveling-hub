package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const sessionPrefix = "travelinghub:web-session:"

type sessionStore struct {
	redis *redis.Client
	ttl   time.Duration
}

func newSessionStore(redisClient *redis.Client, ttl time.Duration) sessionStore {
	return sessionStore{redis: redisClient, ttl: ttl}
}

func (s sessionStore) Create(ctx context.Context, session UserSession) (string, error) {
	id, err := NewSessionID()
	if err != nil {
		return "", err
	}
	value, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("encode session: %w", err)
	}
	if err := s.redis.Set(ctx, sessionKey(id), value, s.ttl).Err(); err != nil {
		return "", fmt.Errorf("store session: %w", err)
	}
	return id, nil
}

func (s sessionStore) Load(ctx context.Context, id string) (UserSession, error) {
	if id == "" {
		return UserSession{}, fmt.Errorf("missing session")
	}
	value, err := s.redis.Get(ctx, sessionKey(id)).Bytes()
	if err != nil {
		return UserSession{}, fmt.Errorf("load session: %w", err)
	}
	var session UserSession
	if err := json.Unmarshal(value, &session); err != nil || session.UserID == uuid.Nil {
		return UserSession{}, fmt.Errorf("decode session")
	}
	return session, nil
}

func (s sessionStore) Delete(ctx context.Context, id string) error {
	if id == "" {
		return nil
	}
	return s.redis.Del(ctx, sessionKey(id)).Err()
}

func sessionKey(id string) string {
	return sessionPrefix + fmt.Sprintf("%x", Digest(id))
}
