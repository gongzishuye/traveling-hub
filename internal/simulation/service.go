package simulation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/faria/traveling-hub/internal/event"
	"github.com/faria/traveling-hub/internal/world"
	"github.com/google/uuid"
)

type Result struct {
	WorldVersion int64     `json:"world_version"`
	EventID      uuid.UUID `json:"event_id"`
}

type Service struct {
	world  world.Service
	events event.Service
}

func NewService(worldService world.Service, eventService event.Service) Service {
	return Service{world: worldService, events: eventService}
}

func (s Service) RunFixtureTick(ctx context.Context, agentID uuid.UUID) (Result, error) {
	state, err := s.world.Advance(ctx)
	if err != nil {
		return Result{}, err
	}
	payload, _ := json.Marshal(map[string]string{
		"kind":    "world_tick_fixture",
		"message": "A quiet hour passed in the shared world.",
	})
	e, err := s.events.Append(ctx, event.Draft{
		Type: "fixture.world_tick", WorldVersion: state.Version, OccurredAt: state.AdvancedAt,
		Payload: payload, Source: "fixture", Audience: []uuid.UUID{agentID},
	})
	if err != nil {
		return Result{}, fmt.Errorf("append fixture event: %w", err)
	}
	return Result{WorldVersion: state.Version, EventID: e.ID}, nil
}
