package identity

import "github.com/google/uuid"

type RegistrationResult struct {
	Created         bool
	AgentID         uuid.UUID
	FrogID          uuid.UUID
	Username        string
	InitialPassword string
	AgentAPIKey     string
	MustChange      bool
}

type UserSession struct {
	UserID         uuid.UUID
	MustChange     bool
	SessionVersion int64
}
